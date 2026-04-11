package notification

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizeProjectedItemDefaults(t *testing.T) {
	t.Parallel()

	item, err := normalizeProjectedItem(ProjectedNotificationInput{
		OrgID:     uuid.New(),
		Title:     "Nuevo insight",
		Body:      "Hay una novedad",
		ProjectID: 42,
		SourceRef: "ins-1",
	})
	if err != nil {
		t.Fatalf("normalizeProjectedItem() error = %v", err)
	}
	if item.Kind != "system" {
		t.Fatalf("expected default kind system, got %q", item.Kind)
	}
	if item.Source != "ponti-backend" {
		t.Fatalf("expected default source ponti-backend, got %q", item.Source)
	}
	if item.RouteHint != "insight_chat" {
		t.Fatalf("expected default route_hint insight_chat, got %q", item.RouteHint)
	}
	if item.NotificationKey == "" {
		t.Fatal("expected notification key to be generated")
	}
}

func TestBuildApprovalNotification(t *testing.T) {
	t.Parallel()

	summary := "Se recomienda revisar el riesgo"
	notification, err := buildApprovalNotification(uuid.MustParse("b4c98b18-4550-4d48-90f2-0b7d8148f83b"), "user@example.com", PendingApproval{
		ID:             "appr-1",
		OrgID:          "b4c98b18-4550-4d48-90f2-0b7d8148f83b",
		RequestID:      "req-1",
		ActionType:     "approve_expense",
		TargetResource: "expense:123",
		Reason:         "Monto alto",
		RiskLevel:      "high",
		Status:         "pending",
		AISummary:      &summary,
		CreatedAt:      "2026-04-09T10:30:00Z",
	})
	if err != nil {
		t.Fatalf("buildApprovalNotification() error = %v", err)
	}
	if notification.ID != "nexus:approval:appr-1" {
		t.Fatalf("unexpected notification ID %q", notification.ID)
	}
	if notification.Kind != "approval" {
		t.Fatalf("unexpected kind %q", notification.Kind)
	}
	if notification.EntityType != "review_approval" {
		t.Fatalf("unexpected entity_type %q", notification.EntityType)
	}
	if notification.RecipientID != "user@example.com" {
		t.Fatalf("unexpected recipient %q", notification.RecipientID)
	}
	if notification.CreatedAt.IsZero() || !notification.CreatedAt.Equal(time.Date(2026, 4, 9, 10, 30, 0, 0, time.UTC)) {
		t.Fatalf("unexpected created_at %v", notification.CreatedAt)
	}
}
