package notification

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	coreinboxdomain "github.com/devpablocristo/core/notifications/go/inbox/usecases/domain"
	"github.com/google/uuid"
)

var allowedNotificationKinds = map[string]struct{}{
	"insight":        {},
	"insight_digest": {},
	"approval":       {},
	"system":         {},
}

var allowedNotificationSources = map[string]struct{}{
	"ponti-ai":      {},
	"ponti-backend": {},
	"nexus":         {},
}

func normalizeProjectedItem(item ProjectedNotificationInput) (ProjectedNotificationInput, error) {
	item.Kind = strings.TrimSpace(item.Kind)
	if item.Kind == "" {
		item.Kind = "system"
	}
	if _, ok := allowedNotificationKinds[item.Kind]; !ok {
		return ProjectedNotificationInput{}, fmt.Errorf("invalid kind: %s", item.Kind)
	}
	item.Source = strings.TrimSpace(item.Source)
	if item.Source == "" {
		item.Source = "ponti-backend"
	}
	if _, ok := allowedNotificationSources[item.Source]; !ok {
		return ProjectedNotificationInput{}, fmt.Errorf("invalid source: %s", item.Source)
	}
	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		return ProjectedNotificationInput{}, fmt.Errorf("title is required")
	}
	item.Body = strings.TrimSpace(item.Body)
	if item.Body == "" {
		return ProjectedNotificationInput{}, fmt.Errorf("body is required")
	}
	item.RouteHint = strings.TrimSpace(item.RouteHint)
	if item.RouteHint == "" {
		item.RouteHint = "insight_chat"
	}
	item.SourceRef = strings.TrimSpace(item.SourceRef)
	item.CreatedBy = strings.TrimSpace(item.CreatedBy)
	if item.CreatedBy == "" {
		item.CreatedBy = "system"
	}
	if item.Payload == nil {
		item.Payload = map[string]any{}
	}
	if item.NotificationKey == "" {
		item.NotificationKey = buildNotificationKey(item)
	}
	return item, nil
}

func buildNotificationKey(item ProjectedNotificationInput) string {
	ref := strings.TrimSpace(item.SourceRef)
	if ref == "" {
		ref = strings.TrimSpace(item.Title)
	}
	if item.ProjectID > 0 {
		return fmt.Sprintf("%s:%s:%d:%s", item.Source, item.Kind, item.ProjectID, ref)
	}
	return fmt.Sprintf("%s:%s:%s", item.Source, item.Kind, ref)
}

type PendingApproval struct {
	ID             string
	OrgID          string
	RequestID      string
	ActionType     string
	TargetResource string
	Reason         string
	RiskLevel      string
	Status         string
	AISummary      *string
	CreatedAt      string
	ExpiresAt      *string
}

func buildApprovalNotification(orgID uuid.UUID, actor string, approval PendingApproval) (coreinboxdomain.Notification, error) {
	approvalID := strings.TrimSpace(approval.ID)
	if orgID == uuid.Nil {
		return coreinboxdomain.Notification{}, fmt.Errorf("org_id is required")
	}
	if strings.TrimSpace(actor) == "" {
		return coreinboxdomain.Notification{}, fmt.Errorf("actor is required")
	}
	if approvalID == "" {
		return coreinboxdomain.Notification{}, fmt.Errorf("approval id is required")
	}

	metadata := map[string]any{
		"source":     "nexus",
		"source_ref": approvalID,
		"route_hint": "insight_chat",
		"severity":   riskLevelSeverity(approval.RiskLevel),
		"created_by": "nexus",
		"approval": map[string]any{
			"id":              approvalID,
			"org_id":          strings.TrimSpace(approval.OrgID),
			"request_id":      strings.TrimSpace(approval.RequestID),
			"action_type":     strings.TrimSpace(approval.ActionType),
			"target_resource": strings.TrimSpace(approval.TargetResource),
			"reason":          strings.TrimSpace(approval.Reason),
			"risk_level":      strings.TrimSpace(approval.RiskLevel),
			"status":          strings.TrimSpace(approval.Status),
			"ai_summary":      approval.AISummary,
			"created_at":      strings.TrimSpace(approval.CreatedAt),
			"expires_at":      approval.ExpiresAt,
		},
	}
	payload, err := json.Marshal(metadata)
	if err != nil {
		return coreinboxdomain.Notification{}, err
	}

	return coreinboxdomain.Notification{
		ID:          approvalNotificationKey(approvalID),
		TenantID:    orgID.String(),
		RecipientID: strings.TrimSpace(actor),
		Title:       buildApprovalTitle(approval),
		Body:        buildApprovalBody(approval),
		Kind:        "approval",
		EntityType:  "review_approval",
		EntityID:    approvalID,
		Metadata:    payload,
		CreatedAt:   parseApprovalTime(approval.CreatedAt),
	}, nil
}

func riskLevelSeverity(riskLevel string) int {
	switch strings.ToLower(strings.TrimSpace(riskLevel)) {
	case "critical":
		return 100
	case "high":
		return 90
	case "medium":
		return 70
	case "low":
		return 40
	default:
		return 50
	}
}

func approvalNotificationKey(approvalID string) string {
	return "nexus:approval:" + strings.TrimSpace(approvalID)
}

func buildApprovalTitle(approval PendingApproval) string {
	actionType := strings.TrimSpace(approval.ActionType)
	target := strings.TrimSpace(approval.TargetResource)
	if actionType == "" {
		actionType = "approval"
	}
	if target == "" {
		return actionType
	}
	return actionType + " - " + target
}

func buildApprovalBody(approval PendingApproval) string {
	reason := strings.TrimSpace(approval.Reason)
	summary := ""
	if approval.AISummary != nil {
		summary = strings.TrimSpace(*approval.AISummary)
	}
	switch {
	case reason != "" && summary != "":
		return reason + "\n\n" + summary
	case reason != "":
		return reason
	case summary != "":
		return summary
	default:
		return "Aprobacion pendiente"
	}
}

func parseApprovalTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return parsed.UTC()
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.UTC()
	}
	return time.Time{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
