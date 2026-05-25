package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	"github.com/devpablocristo/ponti-backend/internal/axis"
)

// CompanionAdapter implementa `usecases.ClientPort` traduciendo las llamadas
// genéricas que el handler hace (`/v1/chat`, `/v1/chat/conversations*`) en
// invocaciones tipadas al cliente Companion.
//
// El handler sigue invocando los métodos sin saber de Companion; el cutover
// vive en el wire y ya no hay cliente ponti-ai legacy.
//
// Notas críticas:
//   - Companion NO tiene SSE: `DoStream` ejecuta el chat síncrono y emite un
//     único evento `done` con el reply completo. El FE actual reconoce `done`
//     como evento final y muestra el texto — UX degradada (sin streaming
//     token-por-token) pero funcional. Trade-off documentado.
//   - `projectID` no se propaga a Companion (Companion no tiene noción de
//     project). Si Ponti necesita filtrar conversaciones por proyecto, hacerlo
//     en el adapter después de obtener la lista (FUTURE: metadata en task).
type CompanionAdapter struct {
	client *axis.CompanionClient
}

// NewCompanionAdapter retorna el adapter listo para usar como `ClientPort`.
func NewCompanionAdapter(client *axis.CompanionClient) *CompanionAdapter {
	return &CompanionAdapter{client: client}
}

// Do enruta la llamada según el path. Los paths se mantienen igual que ponti-ai
// para que el handler no cambie.
func (a *CompanionAdapter) Do(
	ctx context.Context,
	method, urlPath string,
	body any,
	userID, tenantID, projectID string,
) (int, []byte, error) {
	if a.client == nil {
		return 0, nil, domainerr.Unavailable("companion client not configured")
	}
	call := axis.CallContext{
		OrgID: strings.TrimSpace(tenantID),
		Actor: strings.TrimSpace(userID),
	}

	switch {
	case method == http.MethodPost && urlPath == "/v1/chat":
		req, err := toChatRequest(body, projectID)
		if err != nil {
			return 400, nil, domainerr.Validation(err.Error())
		}
		resp, err := a.chatWithOrphanFallback(ctx, call, req)
		if err != nil {
			return mapAxisErr(err)
		}
		return marshalCompanionChat(resp)

	case method == http.MethodGet && urlPath == "/v1/chat/conversations" || strings.HasPrefix(urlPath, "/v1/chat/conversations?"):
		limit := parseLimitQuery(urlPath)
		list, err := a.client.ListConversations(ctx, call, limit)
		if err != nil {
			return mapAxisErr(err)
		}
		return marshalCompanionList(list)

	case method == http.MethodGet && strings.HasPrefix(urlPath, "/v1/chat/conversations/"):
		convID := strings.TrimPrefix(urlPath, "/v1/chat/conversations/")
		convID = path.Clean(convID)
		detail, err := a.client.GetConversation(ctx, call, convID)
		if err != nil {
			return mapAxisErr(err)
		}
		return marshalCompanionDetail(detail)
	}

	return 404, nil, domainerr.NotFound(fmt.Sprintf("companion adapter: route not supported %s %s", method, urlPath))
}

// DoStream simula SSE emitiendo un único evento `done` con el reply completo
// del chat síncrono. El FE actual reconoce `done` y muestra el texto.
//
// Si Companion gana SSE real en el futuro, este método se reescribe para
// proxear el stream sin tocar el handler ni el FE.
func (a *CompanionAdapter) DoStream(
	ctx context.Context,
	method, urlPath string,
	body io.Reader,
	contentType string,
	userID, tenantID, projectID string,
) (*http.Response, error) {
	if a.client == nil {
		return nil, domainerr.Unavailable("companion client not configured")
	}
	if method != http.MethodPost || urlPath != "/v1/chat/stream" {
		return nil, domainerr.NotFound(fmt.Sprintf("companion adapter: stream route not supported %s %s", method, urlPath))
	}

	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read stream body: %w", err)
	}
	var anyBody any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &anyBody); err != nil {
			return nil, domainerr.Validation("invalid stream body json")
		}
	}
	req, err := toChatRequest(anyBody, projectID)
	if err != nil {
		return nil, domainerr.Validation(err.Error())
	}
	call := axis.CallContext{OrgID: strings.TrimSpace(tenantID), Actor: strings.TrimSpace(userID)}
	chat, err := a.chatWithOrphanFallback(ctx, call, req)
	if err != nil {
		return nil, err
	}

	// Componemos un SSE con dos eventos (`start` + `done`) para mantener el
	// contrato actual del FE (`AIAssistant.tsx` espera `done` para mostrar el
	// reply final).
	var sse bytes.Buffer
	startData, _ := json.Marshal(map[string]string{"chat_id": chat.ChatID})
	fmt.Fprintf(&sse, "event: start\ndata: %s\n\n", startData)
	doneData, _ := json.Marshal(map[string]any{
		"chat_id":               chat.ChatID,
		"reply":                 chat.Reply,
		"tool_calls":            chat.ToolCalls,
		"pending_confirmations": chat.PendingConfirmations,
		"blocks":                chat.Blocks,
		"routed_agent":          chat.RoutedAgent,
		"routing_source":        chat.RoutingSource,
	})
	fmt.Fprintf(&sse, "event: done\ndata: %s\n\n", doneData)

	header := make(http.Header)
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("X-Accel-Buffering", "no")

	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     header,
		Body:       io.NopCloser(&sse),
		Request:    nil,
	}, nil
}

// chatWithOrphanFallback ejecuta `Chat` y, si Companion responde 404 NOT_FOUND
// porque el `chat_id` o `task_id` ya no existe en su DB, reintenta sin
// identificadores para crear una conversación nueva. Eso evita que un FE con
// state stale quede atrapado sin poder mandar mensajes.
func (a *CompanionAdapter) chatWithOrphanFallback(ctx context.Context, call axis.CallContext, req axis.ChatRequest) (*axis.ChatResponse, error) {
	resp, err := a.client.Chat(ctx, call, req)
	if err == nil {
		return resp, nil
	}
	// Solo aplica fallback si traíamos un identificador y el error es NOT_FOUND.
	if req.TaskID == "" && req.ChatID == "" {
		return nil, err
	}
	if !errors.Is(err, domainerrNotFound) && !isNotFoundError(err) {
		return nil, err
	}
	// Retry sin identificadores (crea conversación nueva).
	retried := req
	retried.TaskID = ""
	retried.ChatID = ""
	return a.client.Chat(ctx, call, retried)
}

// isNotFoundError detecta errores wrapped que matchean `domainerr.NotFound`
// sin depender de identity exacto del error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	type kindError interface{ Kind() string }
	if ke, ok := err.(kindError); ok {
		return strings.EqualFold(ke.Kind(), "NOT_FOUND")
	}
	return strings.Contains(err.Error(), "NOT_FOUND") || strings.Contains(err.Error(), "not found")
}

// Sentinel para usar errors.Is(...) — domainerr.NotFound es una factory, no un
// valor. Construimos uno con mensaje vacío para comparación por kind.
var domainerrNotFound = domainerr.NotFound("")

// toChatRequest acepta el body genérico que el handler recibe (típicamente un
// `map[string]any` deserializado del JSON del FE) y arma el `ChatRequest` de
// Companion. Tolerante a campos faltantes.
func toChatRequest(body any, projectID string) (axis.ChatRequest, error) {
	if body == nil {
		return axis.ChatRequest{}, errors.New("missing chat body")
	}
	m, ok := body.(map[string]any)
	if !ok {
		// Si viene tipado, re-serializamos y volvemos a parsear como map.
		raw, err := json.Marshal(body)
		if err != nil {
			return axis.ChatRequest{}, fmt.Errorf("marshal body: %w", err)
		}
		if err := json.Unmarshal(raw, &m); err != nil {
			return axis.ChatRequest{}, fmt.Errorf("unmarshal body: %w", err)
		}
	}

	req := axis.ChatRequest{
		ProductSurface: "ponti",
	}
	if v, ok := m["message"].(string); ok {
		req.Message = v
	}
	// El FE manda `chat_id` (UUID de la conversación durable). Companion lo usa
	// para buscar la task asociada en agent_conversations. `task_id` queda sólo
	// para callers internos directos que conocen el UUID de la task.
	if v, ok := m["task_id"].(string); ok && v != "" {
		req.TaskID = v
	} else if v, ok := m["chat_id"].(string); ok {
		req.ChatID = v
	}
	if v, ok := m["channel"].(string); ok && v != "" {
		req.Channel = v
	} else {
		req.Channel = "api"
	}
	// `projectID` queda fuera del payload por ahora: Companion no tiene noción
	// nativa de project en el contrato conversacional.
	_ = projectID
	if req.Message == "" {
		return axis.ChatRequest{}, errors.New("message is required")
	}
	return req, nil
}

// marshalCompanionChat convierte el `axis.ChatResponse` al shape que el handler
// devuelve al FE. Mantenemos los campos del response actual de ponti-ai para
// no romper el FE.
func marshalCompanionChat(r *axis.ChatResponse) (int, []byte, error) {
	out := map[string]any{
		"chat_id":               r.ChatID,
		"reply":                 r.Reply,
		"blocks":                r.Blocks,
		"tool_calls":            r.ToolCalls,
		"pending_confirmations": r.PendingConfirmations,
		"routed_agent":          r.RoutedAgent,
		"routing_source":        r.RoutingSource,
		// Campos legacy de ponti-ai que el FE puede leer; los completamos con
		// defaults razonables porque Companion no los devuelve.
		"request_id":       r.Task.ID,
		"output_kind":      "chat_reply",
		"content_language": "es",
		"tokens_used":      0,
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return 0, nil, fmt.Errorf("marshal chat response: %w", err)
	}
	return 200, raw, nil
}

// marshalCompanionList convierte la lista de conversaciones al shape del FE.
func marshalCompanionList(list *axis.ConversationList) (int, []byte, error) {
	out := map[string]any{"items": list.Items}
	raw, err := json.Marshal(out)
	if err != nil {
		return 0, nil, fmt.Errorf("marshal conversations list: %w", err)
	}
	return 200, raw, nil
}

// marshalCompanionDetail convierte el detalle al shape del FE.
func marshalCompanionDetail(detail *axis.ConversationDetail) (int, []byte, error) {
	type frontendMessage struct {
		Role      string           `json:"role"`
		Content   string           `json:"content"`
		Timestamp *time.Time       `json:"ts,omitempty"`
		Blocks    []axis.ChatBlock `json:"blocks,omitempty"`
	}
	type frontendDetail struct {
		ID        string            `json:"id"`
		Title     string            `json:"title"`
		Messages  []frontendMessage `json:"messages"`
		CreatedAt time.Time         `json:"created_at"`
		UpdatedAt time.Time         `json:"updated_at"`
	}

	messages := make([]frontendMessage, 0, len(detail.Messages))
	for _, msg := range detail.Messages {
		var ts *time.Time
		if !msg.Timestamp.IsZero() {
			t := msg.Timestamp
			ts = &t
		}
		messages = append(messages, frontendMessage{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: ts,
			Blocks:    msg.Blocks,
		})
	}

	raw, err := json.Marshal(frontendDetail{
		ID:        detail.ID,
		Title:     detail.Title,
		Messages:  messages,
		CreatedAt: detail.CreatedAt,
		UpdatedAt: detail.UpdatedAt,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("marshal conversation detail: %w", err)
	}
	return 200, raw, nil
}

// mapAxisErr traduce errores tipados del cliente axis a (status, body, err)
// para que el handler maneje el caso fallback dummy igual que con ponti-ai.
func mapAxisErr(err error) (int, []byte, error) {
	if errors.Is(err, axis.ErrNotConfigured) {
		return 0, nil, domainerr.Unavailable("companion client not configured")
	}
	return 0, nil, err
}

// parseLimitQuery extrae el `?limit=N` del path para reenviarlo a Companion.
func parseLimitQuery(urlPath string) int {
	idx := strings.Index(urlPath, "?")
	if idx < 0 {
		return 0
	}
	for _, kv := range strings.Split(urlPath[idx+1:], "&") {
		if strings.HasPrefix(kv, "limit=") {
			n, err := strconv.Atoi(strings.TrimPrefix(kv, "limit="))
			if err == nil {
				return n
			}
		}
	}
	return 0
}
