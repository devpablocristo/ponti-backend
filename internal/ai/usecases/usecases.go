// Package usecases contiene casos de uso del proxy AI.
package usecases

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

type ClientPort interface {
	Do(ctx context.Context, method, path string, body any, userID, projectID string) (int, []byte, error)
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

func (u *UseCases) ComputeInsights(ctx context.Context, userID, projectID string) (int, []byte, error) {
	return u.dummyOrReal(ctx, "POST", "/v1/insights/compute", nil, userID, projectID, map[string]any{
		"request_id":       "dummy",
		"computed":         0,
		"insights_created": 0,
	})
}

func (u *UseCases) GetInsights(ctx context.Context, userID, projectID, entityType, entityID string) (int, []byte, error) {
	path := "/v1/insights/" + entityType + "/" + entityID
	return u.dummyOrReal(ctx, "GET", path, nil, userID, projectID, map[string]any{
		"insights": []any{},
	})
}

func (u *UseCases) GetSummary(ctx context.Context, userID, projectID string) (int, []byte, error) {
	return u.dummyOrReal(ctx, "GET", "/v1/insights/summary", nil, userID, projectID, map[string]any{
		"new_count_total":         0,
		"new_count_high_severity": 0,
		"top_insights":            []any{},
	})
}

func (u *UseCases) ExplainInsight(
	ctx context.Context,
	userID, projectID, insightID, mode string,
) (int, []byte, error) {
	path := "/v1/copilot/insights/" + insightID + "/" + mode
	return u.dummyOrReal(ctx, "GET", path, nil, userID, projectID, map[string]any{
		"insight_id": insightID,
		"mode":       mode,
		"explanation": map[string]any{
			"human_readable":     "AI copilot no configurado",
			"audit_focused":      "AI copilot no configurado",
			"what_to_watch_next": "AI copilot no configurado",
		},
		"proposal": nil,
	})
}

func (u *UseCases) RecordAction(ctx context.Context, userID, projectID, insightID string, body any) (int, []byte, error) {
	path := "/v1/insights/" + insightID + "/actions"
	return u.dummyOrReal(ctx, "POST", path, body, userID, projectID, map[string]any{
		"request_id": "dummy",
		"status":     "dummy",
	})
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
