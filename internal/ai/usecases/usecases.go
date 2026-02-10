// Package usecases contiene casos de uso del proxy AI.
package usecases

import (
	"context"
	"encoding/json"
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

func (u *UseCases) Ask(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.dummyOrReal(ctx, "POST", "/v1/ask", body, userID, projectID, map[string]any{
		"request_id": "dummy",
		"intent":     "placeholder",
		"data":       []any{},
		"answer":     "AI no configurada. Configurar AI_SERVICE_URL y AI_SERVICE_KEY en Cloud Run.",
		"sources":    []any{},
		"warnings":   []string{"AI service not configured"},
	})
}

func (u *UseCases) Ingest(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.dummyOrReal(ctx, "POST", "/v1/rag/ingest", body, userID, projectID, map[string]any{
		"request_id": "dummy",
		"ingested":   0,
	})
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
		"new_count_total":        0,
		"new_count_high_severity": 0,
		"top_insights":            []any{},
	})
}

func (u *UseCases) RecordAction(ctx context.Context, userID, projectID, insightID string, body any) (int, []byte, error) {
	path := "/v1/insights/" + insightID + "/actions"
	return u.dummyOrReal(ctx, "POST", path, body, userID, projectID, map[string]any{
		"request_id": "dummy",
		"status":     "dummy",
	})
}

func (u *UseCases) RecomputeActive(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.dummyOrReal(ctx, "POST", "/v1/jobs/recompute-active", body, userID, projectID, map[string]any{
		"request_id": "dummy",
		"status":     "dummy",
	})
}

func (u *UseCases) RecomputeBaselines(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.dummyOrReal(ctx, "POST", "/v1/jobs/recompute-baselines", body, userID, projectID, map[string]any{
		"request_id": "dummy",
		"status":     "dummy",
	})
}
