// Package usecases contiene los casos de uso del proxy AI.
//
// La ruta activa es Axis Companion, inyectado vía `ai.CompanionAdapter` que
// implementa `ClientPort`.
package usecases

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ClientPort es la interfaz que el handler usa para hablar con el upstream AI.
// La implementa `ai.CompanionAdapter` traduciendo cada llamada al cliente
// tipado de Companion en `internal/axis`.
type ClientPort interface {
	Do(ctx context.Context, method, path string, body any, userID, tenantID, projectID string) (int, []byte, error)
	DoStream(ctx context.Context, method, path string, body io.Reader, contentType string, userID, tenantID, projectID string) (*http.Response, error)
}

// UseCases orquesta el chat conversacional contra Companion. No tiene lógica
// extra hoy — sólo route + clamp del limit en listados — pero queda como
// punto de extensión si después se agrega cache, métricas custom, retries
// específicos, o el cliente Nexus para gating.
type UseCases struct {
	client ClientPort
}

// NewUseCases construye los casos de uso. `client` debe estar ya configurado
// (sin URL no se construye y el binary falla en wire).
func NewUseCases(client ClientPort) *UseCases {
	return &UseCases{client: client}
}

// ChatStream proxea POST /v1/chat/stream. El adapter compone SSE sintético
// (start + done) sobre el response síncrono de Companion — el FE no nota la
// diferencia salvo por la falta de tokens progresivos.
func (u *UseCases) ChatStream(ctx context.Context, userID, tenantID, projectID string, body io.Reader, w http.ResponseWriter) error {
	resp, err := u.client.DoStream(ctx, http.MethodPost, "/v1/chat/stream", body, "application/json", userID, tenantID, projectID)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func (u *UseCases) Chat(ctx context.Context, userID, tenantID, projectID string, body any) (int, []byte, error) {
	return u.client.Do(ctx, "POST", "/v1/chat", body, userID, tenantID, projectID)
}

func (u *UseCases) ListChatConversations(ctx context.Context, userID, tenantID, projectID string, limit int) (int, []byte, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	path := "/v1/chat/conversations?limit=" + strconv.Itoa(limit)
	return u.client.Do(ctx, "GET", path, nil, userID, tenantID, projectID)
}

func (u *UseCases) GetChatConversation(ctx context.Context, userID, tenantID, projectID, conversationID string) (int, []byte, error) {
	path := "/v1/chat/conversations/" + strings.TrimSpace(conversationID)
	return u.client.Do(ctx, "GET", path, nil, userID, tenantID, projectID)
}
