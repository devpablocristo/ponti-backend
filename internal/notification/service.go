package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type summaryEnvelope struct {
	TopInsights []summaryInsight `json:"top_insights"`
}

type summaryInsight struct {
	ID         string `json:"id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Type       string `json:"type"`
	Severity   int    `json:"severity"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Status     string `json:"status"`
	DedupeKey  string `json:"dedupe_key"`
}

type InsightSyncService struct {
	repo *Repository
}

func NewInsightSyncService(repo *Repository) *InsightSyncService {
	return &InsightSyncService{repo: repo}
}

func (s *InsightSyncService) SyncFromSummary(ctx context.Context, orgID uuid.UUID, projectID int64, actor string, raw []byte) error {
	if orgID == uuid.Nil || projectID <= 0 || len(raw) == 0 {
		return nil
	}
	var summary summaryEnvelope
	if err := json.Unmarshal(raw, &summary); err != nil {
		return err
	}
	items := make([]ProjectedNotificationInput, 0, len(summary.TopInsights))
	for _, insight := range summary.TopInsights {
		if strings.TrimSpace(insight.ID) == "" {
			continue
		}
		key := strings.TrimSpace(insight.DedupeKey)
		if key == "" {
			key = fmt.Sprintf("insight:%d:%s", projectID, insight.ID)
		}
		items = append(items, ProjectedNotificationInput{
			OrgID:           orgID,
			ProjectID:       projectID,
			Actor:           actor,
			Kind:            "insight",
			Source:          "ponti-ai",
			SourceRef:       insight.ID,
			NotificationKey: key,
			Title:           strings.TrimSpace(insight.Title),
			Body:            strings.TrimSpace(insight.Summary),
			Severity:        insight.Severity,
			RouteHint:       "insight_chat",
			CreatedBy:       defaultString(actor, "system"),
			Payload: map[string]any{
				"insight_id":   insight.ID,
				"entity_type":  insight.EntityType,
				"entity_id":    insight.EntityID,
				"insight_type": insight.Type,
				"route_hint":   "insight_chat",
				"source":       "insights.summary",
				"status":       insight.Status,
				"chat_context": map[string]any{
					"insight_id":             insight.ID,
					"scope":                  fmt.Sprintf("%s:%s", insight.EntityType, insight.EntityID),
					"routed_agent":           "insight_chat",
					"content_language":       "es",
					"suggested_user_message": fmt.Sprintf("Explicame este insight: %s", insight.Title),
					"source_kind":            "insight",
				},
			},
		})
	}
	if len(items) == 0 {
		return nil
	}
	return s.repo.UpsertProjectedNotifications(ctx, items)
}

type ProjectionService struct {
	repo *Repository
}

func NewProjectionService(repo *Repository) *ProjectionService {
	return &ProjectionService{repo: repo}
}

func (s *ProjectionService) Project(ctx context.Context, items []ProjectedNotificationInput) error {
	if len(items) == 0 {
		return nil
	}
	normalized := make([]ProjectedNotificationInput, 0, len(items))
	for _, item := range items {
		clean, err := normalizeProjectedItem(item)
		if err != nil {
			return err
		}
		normalized = append(normalized, clean)
	}
	return s.repo.UpsertProjectedNotifications(ctx, normalized)
}
