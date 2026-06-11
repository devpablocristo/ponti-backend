// Package usecases contiene casos de uso del proxy AI (chat con copilot).
package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/axis"
)

type ClientPort interface {
	Do(ctx context.Context, method, path string, body any, userID, projectID string) (int, []byte, error)
	DoStream(ctx context.Context, method, path string, body io.Reader, contentType string, userID, projectID string) (*http.Response, error)
}

type AxisClientPort interface {
	DoJSON(ctx context.Context, call axis.CallContext, method, path string, body any) (int, []byte, error)
	ProductSurface() string
}

type Config struct {
	Provider       string
	AxisEnabled    bool
	ProductSurface string
}

type UseCases struct {
	client     ClientPort
	axisClient AxisClientPort
	cfg        Config
}

func NewUseCases(client ClientPort, axisClient AxisClientPort, cfg Config) *UseCases {
	if strings.TrimSpace(cfg.ProductSurface) == "" {
		cfg.ProductSurface = "ponti"
	}
	if strings.TrimSpace(cfg.Provider) == "" {
		cfg.Provider = "legacy"
	}
	return &UseCases{client: client, axisClient: axisClient, cfg: cfg}
}

// isAIServiceNotConfigured indica si el error es por AI no configurada (URL/KEY vacíos).
func isAIServiceNotConfigured(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "ai service url not configured") ||
		strings.Contains(s, "ai service key not configured")
}

func (u *UseCases) axisActive() bool {
	return u.cfg.AxisEnabled && strings.EqualFold(strings.TrimSpace(u.cfg.Provider), "axis")
}

// dummyOrReal ejecuta la llamada al cliente; si AI no está configurada, devuelve respuestas dummy.
func (u *UseCases) dummyOrReal(ctx context.Context, method, path string, body any, userID, projectID string, dummyResp any) (int, []byte, error) {
	status, raw, err := u.client.Do(ctx, method, path, body, userID, projectID)
	if err == nil {
		return status, raw, nil
	}
	if !isAIServiceNotConfigured(err) {
		return 0, nil, err
	}
	b, _ := json.Marshal(dummyResp)
	return http.StatusOK, b, nil
}

// ChatStream proxea SSE hacia ponti-ai en modo legacy. En modo Axis sintetiza
// eventos compatibles para no romper la UI web mientras Companion estabiliza streaming.
func (u *UseCases) ChatStream(ctx context.Context, userID, projectID string, body io.Reader, w http.ResponseWriter) error {
	if u.axisActive() {
		return u.chatStreamAxis(ctx, userID, projectID, body, w)
	}
	resp, err := u.client.DoStream(ctx, http.MethodPost, "/v1/chat/stream", body, "application/json", userID, projectID)
	if err != nil {
		if isAIServiceNotConfigured(err) {
			w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("X-Accel-Buffering", "no")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, "event: error\ndata: {\"message\":\"ai_not_configured\"}\n\n")
			return nil
		}
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

func (u *UseCases) Chat(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	dummy := map[string]any{
		"request_id":            "dummy",
		"output_kind":           "chat_reply",
		"content_language":      "es",
		"chat_id":               "",
		"reply":                 "Asistente AI no configurado.",
		"tokens_used":           0,
		"tool_calls":            []any{},
		"pending_confirmations": []any{},
		"blocks":                []any{},
		"routed_agent":          "general",
		"routing_source":        "read_fallback",
	}
	if !u.axisActive() {
		slog.InfoContext(ctx, "ponti_ai_chat_provider",
			"provider", "legacy",
			"fallback_used", false,
			"tenant_id", tenantIDFromContext(ctx),
			"project_id", strings.TrimSpace(projectID),
			"product_surface", u.productSurface(),
		)
		return u.dummyOrReal(ctx, http.MethodPost, "/v1/chat", body, userID, projectID, dummy)
	}
	status, raw, err := u.chatAxis(ctx, userID, projectID, body)
	if err == nil && status < http.StatusInternalServerError {
		return status, raw, nil
	}
	if err != nil {
		slog.WarnContext(ctx, "ponti_ai_axis_fallback",
			"provider", "axis",
			"fallback_used", true,
			"fallback_reason", "axis_request_error",
			"error", err.Error(),
			"tenant_id", tenantIDFromContext(ctx),
			"project_id", strings.TrimSpace(projectID),
			"product_surface", u.productSurface(),
		)
		return u.dummyOrReal(ctx, http.MethodPost, "/v1/chat", body, userID, projectID, dummy)
	}
	if status >= http.StatusInternalServerError {
		slog.WarnContext(ctx, "ponti_ai_axis_fallback",
			"provider", "axis",
			"axis_status", status,
			"fallback_used", true,
			"fallback_reason", "axis_server_status",
			"tenant_id", tenantIDFromContext(ctx),
			"project_id", strings.TrimSpace(projectID),
			"product_surface", u.productSurface(),
		)
		return u.dummyOrReal(ctx, http.MethodPost, "/v1/chat", body, userID, projectID, dummy)
	}
	return status, raw, err
}

func (u *UseCases) ListChatConversations(ctx context.Context, userID, projectID string, limit int) (int, []byte, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	path := "/v1/chat/conversations?limit=" + strconv.Itoa(limit)
	dummy := map[string]any{"items": []any{}}
	if !u.axisActive() {
		return u.dummyOrReal(ctx, http.MethodGet, path, nil, userID, projectID, dummy)
	}
	status, raw, err := u.axisClient.DoJSON(ctx, u.axisCallContext(ctx, userID, true), http.MethodGet, path, nil)
	if err == nil && status < http.StatusInternalServerError {
		return status, raw, nil
	}
	return u.dummyOrReal(ctx, http.MethodGet, path, nil, userID, projectID, dummy)
}

func (u *UseCases) GetChatConversation(ctx context.Context, userID, projectID, conversationID string) (int, []byte, error) {
	path := "/v1/chat/conversations/" + strings.TrimSpace(conversationID)
	dummy := map[string]any{
		"id":         conversationID,
		"title":      "dummy",
		"messages":   []any{},
		"created_at": "",
		"updated_at": "",
	}
	if !u.axisActive() {
		return u.dummyOrReal(ctx, http.MethodGet, path, nil, userID, projectID, dummy)
	}
	status, raw, err := u.axisClient.DoJSON(ctx, u.axisCallContext(ctx, userID, true), http.MethodGet, path, nil)
	if err == nil && status < http.StatusInternalServerError {
		return status, raw, nil
	}
	return u.dummyOrReal(ctx, http.MethodGet, path, nil, userID, projectID, dummy)
}

func (u *UseCases) chatAxis(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	payload, err := normalizeMap(body)
	if err != nil {
		return http.StatusBadRequest, jsonError("VALIDATION", "invalid request payload"), nil
	}
	if tenantID := tenantIDFromContext(ctx); tenantID == "" {
		return http.StatusBadRequest, jsonError("VALIDATION", "tenant context is required"), nil
	}
	if strings.TrimSpace(projectID) == "" {
		return http.StatusBadRequest, jsonError("VALIDATION", "project_id is required"), nil
	}
	axisReq := map[string]any{
		"message":         payload["message"],
		"channel":         "web",
		"product_surface": u.productSurface(),
	}
	copyIfPresent(axisReq, payload, "chat_id")
	copyIfPresent(axisReq, payload, "route_hint")
	copyIfPresent(axisReq, payload, "confirmed_actions")
	copyIfPresent(axisReq, payload, "agent_id")
	if handoff := buildHandoff(payload); handoff != nil {
		axisReq["handoff"] = handoff
	}
	status, raw, err := u.axisClient.DoJSON(ctx, u.axisCallContext(ctx, userID, false), http.MethodPost, "/v1/chat", axisReq)
	if err != nil {
		return status, raw, err
	}
	runID, taskID := extractAxisIDs(raw)
	slog.InfoContext(ctx, "ponti_ai_axis_chat_response",
		"provider", "axis",
		"axis_status", status,
		"fallback_used", false,
		"axis_run_id", runID,
		"axis_task_id", taskID,
		"tenant_id", tenantIDFromContext(ctx),
		"project_id", strings.TrimSpace(projectID),
		"product_surface", u.productSurface(),
	)
	if status >= http.StatusBadRequest {
		return status, raw, nil
	}
	adapted, err := adaptAxisChatResponse(raw, payload)
	if err != nil {
		return status, raw, nil
	}
	return status, adapted, nil
}

func (u *UseCases) chatStreamAxis(ctx context.Context, userID, projectID string, body io.Reader, w http.ResponseWriter) error {
	rawBody, err := io.ReadAll(io.LimitReader(body, 1<<20))
	if err != nil {
		return err
	}
	var payload map[string]any
	if len(bytes.TrimSpace(rawBody)) == 0 {
		payload = map[string]any{}
	} else if err := json.Unmarshal(rawBody, &payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(jsonError("VALIDATION", "invalid request payload"))
		return nil
	}
	status, raw, err := u.Chat(ctx, userID, projectID, payload)
	if err != nil {
		return err
	}
	if status >= http.StatusBadRequest {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(raw)
		return nil
	}
	var response map[string]any
	if err := json.Unmarshal(raw, &response); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	writeSSE(w, "start", map[string]any{"chat_id": response["chat_id"]})
	for _, tool := range toolCallObjects(response["tool_calls"]) {
		writeSSE(w, "tool_call", map[string]any{"tool": tool})
	}
	if reply, _ := response["reply"].(string); reply != "" {
		writeSSE(w, "text", map[string]any{"content": reply})
	}
	writeSSE(w, "done", response)
	return nil
}

func (u *UseCases) axisCallContext(ctx context.Context, userID string, readOnly bool) axis.CallContext {
	scopes := scopesFromContext(ctx)
	scopes = append(scopes, "ponti:insights:read", "ponti:operational:read")
	if readOnly {
		scopes = append(scopes, "companion:tasks:read")
	} else {
		scopes = append(scopes, "companion:tasks:write", "companion:connectors:execute", "ponti:actions:draft")
	}
	return axis.CallContext{
		OrgID:          tenantIDFromContext(ctx),
		ActorID:        strings.TrimSpace(userID),
		OnBehalfOf:     strings.TrimSpace(userID),
		ProductSurface: u.productSurface(),
		Scopes:         scopes,
	}
}

func (u *UseCases) productSurface() string {
	if strings.TrimSpace(u.cfg.ProductSurface) != "" {
		return strings.TrimSpace(u.cfg.ProductSurface)
	}
	if u.axisClient != nil {
		return u.axisClient.ProductSurface()
	}
	return "ponti"
}

func tenantIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	switch v := ctx.Value(ctxkeys.OrgID).(type) {
	case uuid.UUID:
		if v == uuid.Nil {
			return ""
		}
		return v.String()
	case string:
		return strings.TrimSpace(v)
	default:
		return ""
	}
}

func scopesFromContext(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	if scopes, ok := ctx.Value(ctxkeys.Scopes).([]string); ok {
		return append([]string{}, scopes...)
	}
	return nil
}

func normalizeMap(body any) (map[string]any, error) {
	if body == nil {
		return map[string]any{}, nil
	}
	if m, ok := body.(map[string]any); ok {
		return m, nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func copyIfPresent(dst, src map[string]any, key string) {
	if v, ok := src[key]; ok && v != nil {
		dst[key] = v
	}
}

func buildHandoff(payload map[string]any) any {
	if v, ok := payload["handoff"]; ok && v != nil {
		return v
	}
	routeHint, _ := payload["route_hint"].(string)
	if workspace, ok := payload["workspace"]; ok && workspace != nil {
		out := map[string]any{
			"source":    "ponti-web",
			"workspace": workspace,
		}
		if strings.TrimSpace(routeHint) != "" {
			out["route_hint"] = strings.TrimSpace(routeHint)
		}
		return out
	}
	if strings.TrimSpace(routeHint) != "" {
		return map[string]any{
			"source":     "ponti-web",
			"route_hint": strings.TrimSpace(routeHint),
		}
	}
	return nil
}

func adaptAxisChatResponse(raw []byte, request map[string]any) ([]byte, error) {
	var in map[string]any
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, err
	}
	chatID := stringValue(in["chat_id"])
	if chatID == "" {
		chatID = stringValue(request["chat_id"])
	}
	taskID := stringValue(in["task_id"])
	runID := stringValue(in["run_id"])
	requestID := firstNonEmpty(runID, taskID, chatID, stringValue(in["request_id"]))
	reply := stringValue(in["reply"])
	blocks := arrayValue(in["blocks"])
	if len(blocks) == 0 && reply != "" {
		blocks = []any{map[string]any{"type": "text", "text": reply}}
	}
	routedAgent := firstNonEmpty(stringValue(in["routed_agent"]), stringValue(in["agent_id"]), "companion")
	routingSource := firstNonEmpty(stringValue(in["routing_source"]), "axis")
	out := map[string]any{
		"request_id":            requestID,
		"output_kind":           firstNonEmpty(stringValue(in["output_kind"]), "chat_reply"),
		"content_language":      firstNonEmpty(stringValue(request["preferred_language"]), "es"),
		"chat_id":               chatID,
		"reply":                 reply,
		"tokens_used":           intValue(in["tokens_used"]),
		"tool_calls":            toolCallObjects(in["tool_calls"]),
		"pending_confirmations": arrayValue(in["pending_confirmations"]),
		"blocks":                blocks,
		"routed_agent":          routedAgent,
		"routing_source":        routingSource,
		"axis_run_id":           runID,
		"axis_task_id":          taskID,
		"run_id":                runID,
		"task_id":               taskID,
		"agent_id":              stringValue(in["agent_id"]),
	}
	return json.Marshal(out)
}

func extractAxisIDs(raw []byte) (runID, taskID string) {
	var in map[string]any
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", ""
	}
	return stringValue(in["run_id"]), stringValue(in["task_id"])
}

func jsonError(code, message string) []byte {
	raw, _ := json.Marshal(map[string]string{"code": code, "message": message})
	return raw
}

func writeSSE(w http.ResponseWriter, event string, data any) {
	raw, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, raw)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func stringValue(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return ""
	}
}

func intValue(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

func arrayValue(v any) []any {
	if arr, ok := v.([]any); ok {
		return arr
	}
	return []any{}
}

func toolCallObjects(v any) []any {
	items := []any{}
	for _, item := range arrayValue(v) {
		switch t := item.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" {
				items = append(items, s)
			}
		case map[string]any:
			items = append(items, t)
		}
	}
	return items
}

func toolCallNames(v any) []string {
	out := []string{}
	for _, item := range arrayValue(v) {
		switch t := item.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" {
				out = append(out, s)
			}
		case map[string]any:
			if s := firstNonEmpty(stringValue(t["name"]), stringValue(t["tool"]), stringValue(t["capability_id"])); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
