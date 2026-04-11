package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	coredomain "github.com/devpablocristo/core/notifications/go/inbox/usecases/domain"
	"github.com/google/uuid"
)

func (r *Repository) ListForRecipient(ctx context.Context, tenantID, recipientID string, limit int) ([]coredomain.Notification, error) {
	orgID, err := uuid.Parse(strings.TrimSpace(tenantID))
	if err != nil {
		return nil, err
	}
	items, err := r.List(ctx, ListFilters{
		OrgID: orgID,
		Actor: strings.TrimSpace(recipientID),
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]coredomain.Notification, 0, len(items))
	for _, item := range items {
		metadata, err := json.Marshal(item.Payload)
		if err != nil {
			return nil, err
		}
		out = append(out, coredomain.Notification{
			ID:          item.NotificationKey,
			TenantID:    tenantID,
			RecipientID: strings.TrimSpace(recipientID),
			Title:       item.Title,
			Body:        item.Body,
			Kind:        item.Kind,
			EntityType:  metadataString(item.Payload, "entity_type"),
			EntityID:    metadataString(item.Payload, "entity_id"),
			Metadata:    metadata,
			ReadAt:      item.ReadAt,
			CreatedAt:   item.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repository) CountUnread(ctx context.Context, tenantID, recipientID string) (int64, error) {
	orgID, err := uuid.Parse(strings.TrimSpace(tenantID))
	if err != nil {
		return 0, err
	}
	summary, err := r.GetSummary(ctx, orgID, strings.TrimSpace(recipientID), nil)
	if err != nil {
		return 0, err
	}
	return summary.NewCount, nil
}

func (r *Repository) Append(ctx context.Context, notification coredomain.Notification) (coredomain.Notification, error) {
	orgID, err := uuid.Parse(strings.TrimSpace(notification.TenantID))
	if err != nil {
		return coredomain.Notification{}, err
	}
	payload := map[string]any{}
	if len(notification.Metadata) > 0 {
		if err := json.Unmarshal(notification.Metadata, &payload); err != nil {
			return coredomain.Notification{}, err
		}
	}
	projectID := metadataInt64(payload, "project_id")
	input, err := normalizeProjectedItem(ProjectedNotificationInput{
		OrgID:           orgID,
		ProjectID:       projectID,
		Actor:           strings.TrimSpace(notification.RecipientID),
		Kind:            firstNonEmpty(strings.TrimSpace(notification.Kind), "system"),
		Source:          firstNonEmpty(metadataString(payload, "source"), "ponti-backend"),
		SourceRef:       firstNonEmpty(metadataString(payload, "source_ref"), notification.EntityID),
		NotificationKey: firstNonEmpty(strings.TrimSpace(notification.ID), buildNotificationKey(ProjectedNotificationInput{Source: "ponti-backend", Kind: notification.Kind, SourceRef: notification.EntityID})),
		Title:           notification.Title,
		Body:            notification.Body,
		Severity:        metadataInt(payload, "severity"),
		RouteHint:       metadataString(payload, "route_hint"),
		CreatedBy:       firstNonEmpty(metadataString(payload, "created_by"), "system"),
		Payload:         payload,
	})
	if err != nil {
		return coredomain.Notification{}, err
	}
	if err := r.UpsertProjectedNotifications(ctx, []ProjectedNotificationInput{input}); err != nil {
		return coredomain.Notification{}, err
	}
	notification.ID = input.NotificationKey
	return notification, nil
}

func (r *Repository) MarkRead(ctx context.Context, tenantID, recipientID, notificationID string, readAt time.Time) (time.Time, error) {
	orgID, err := uuid.Parse(strings.TrimSpace(tenantID))
	if err != nil {
		return time.Time{}, err
	}
	actor := strings.TrimSpace(recipientID)
	now := readAt.UTC()
	res := r.baseInboxQuery(ctx, orgID, actor, nil).
		Where("notification_key = ?", strings.TrimSpace(notificationID)).
		Updates(map[string]any{
			"status":     "read",
			"read_at":    now,
			"updated_at": now,
		})
	if res.Error != nil {
		return time.Time{}, res.Error
	}
	if res.RowsAffected == 0 {
		return time.Time{}, fmt.Errorf("notification not found")
	}
	return now, nil
}

func metadataString(payload map[string]any, key string) string {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func metadataInt(payload map[string]any, key string) int {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return 0
	}
	switch value := raw.(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func metadataInt64(payload map[string]any, key string) int64 {
	raw, ok := payload[key]
	if !ok || raw == nil {
		return 0
	}
	switch value := raw.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}
