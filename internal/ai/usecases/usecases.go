// Package usecases contiene casos de uso del proxy AI (chat con copilot).
package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ClientPort interface {
	Do(ctx context.Context, method, path string, body any, userID, projectID string) (int, []byte, error)
	DoStream(ctx context.Context, method, path string, body io.Reader, contentType string, userID, projectID string) (*http.Response, error)
}

type UseCases struct {
	client ClientPort
}

func NewUseCases(client ClientPort) *UseCases {
	return &UseCases{client: client}
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
	return 200, b, nil
}

// ChatStream proxea POST /v1/chat/stream hacia ponti-ai (SSE); escribe headers y cuerpo en w.
func (u *UseCases) ChatStream(ctx context.Context, userID, projectID string, body io.Reader, w http.ResponseWriter) error {
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
	return u.dummyOrReal(ctx, "POST", "/v1/chat", body, userID, projectID, map[string]any{
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
	})
}

func (u *UseCases) ListChatConversations(ctx context.Context, userID, projectID string, limit int) (int, []byte, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	path := "/v1/chat/conversations?limit=" + strconv.Itoa(limit)
	return u.dummyOrReal(ctx, "GET", path, nil, userID, projectID, map[string]any{
		"items": []any{},
	})
}

func (u *UseCases) GetChatConversation(ctx context.Context, userID, projectID, conversationID string) (int, []byte, error) {
	path := "/v1/chat/conversations/" + strings.TrimSpace(conversationID)
	return u.dummyOrReal(ctx, "GET", path, nil, userID, projectID, map[string]any{
		"id":         conversationID,
		"title":      "dummy",
		"messages":   []any{},
		"created_at": "",
		"updated_at": "",
	})
}
