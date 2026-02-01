// Package usecases contiene casos de uso del proxy AI.
package usecases

import (
	"context"
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

func (u *UseCases) Ask(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.client.Do(ctx, "POST", "/v1/ask", body, userID, projectID)
}

func (u *UseCases) Ingest(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.client.Do(ctx, "POST", "/v1/rag/ingest", body, userID, projectID)
}

func (u *UseCases) ComputeInsights(ctx context.Context, userID, projectID string) (int, []byte, error) {
	return u.client.Do(ctx, "POST", "/v1/insights/compute", nil, userID, projectID)
}

func (u *UseCases) GetInsights(ctx context.Context, userID, projectID, entityType, entityID string) (int, []byte, error) {
	path := "/v1/insights/" + entityType + "/" + entityID
	return u.client.Do(ctx, "GET", path, nil, userID, projectID)
}

func (u *UseCases) GetSummary(ctx context.Context, userID, projectID string) (int, []byte, error) {
	return u.client.Do(ctx, "GET", "/v1/insights/summary", nil, userID, projectID)
}

func (u *UseCases) RecordAction(ctx context.Context, userID, projectID, insightID string, body any) (int, []byte, error) {
	path := "/v1/insights/" + insightID + "/actions"
	return u.client.Do(ctx, "POST", path, body, userID, projectID)
}

func (u *UseCases) RecomputeActive(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.client.Do(ctx, "POST", "/v1/jobs/recompute-active", body, userID, projectID)
}

func (u *UseCases) RecomputeBaselines(ctx context.Context, userID, projectID string, body any) (int, []byte, error) {
	return u.client.Do(ctx, "POST", "/v1/jobs/recompute-baselines", body, userID, projectID)
}
