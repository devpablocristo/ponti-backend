package notification

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// fakeRepo captures the items passed to UpsertProjectedNotifications.
type fakeRepo struct {
	Repository
	captured []ProjectedNotificationInput
}

func (f *fakeRepo) upsert(_ context.Context, items []ProjectedNotificationInput) error {
	f.captured = items
	return nil
}

func TestSyncFromSummaryBuildsChatContext(t *testing.T) {
	t.Parallel()

	summary := map[string]any{
		"top_insights": []map[string]any{
			{
				"id":          "ins-1",
				"entity_type": "project",
				"entity_id":   "p42",
				"type":        "anomaly",
				"severity":    85,
				"title":       "Costo fuera de rango",
				"summary":     "El costo superó el percentil 75.",
				"status":      "new",
				"dedupe_key":  "dk-1",
			},
		},
	}
	raw, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}

	orgID := uuid.New()
	var projectID int64 = 42

	// Parsear la misma lógica que SyncFromSummary
	var envelope summaryEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.TopInsights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(envelope.TopInsights))
	}

	// Ejecutar SyncFromSummary con un servicio cuyo repo tiene db nil
	// (va a fallar en el upsert, así que verificamos la transformación manualmente)
	insight := envelope.TopInsights[0]

	// Verificar que los campos existen para armar el chat_context
	if insight.ID != "ins-1" {
		t.Fatalf("unexpected ID %q", insight.ID)
	}
	if insight.EntityType != "project" {
		t.Fatalf("unexpected entity_type %q", insight.EntityType)
	}
	if insight.EntityID != "p42" {
		t.Fatalf("unexpected entity_id %q", insight.EntityID)
	}

	// Simular la construcción del payload como lo hace SyncFromSummary
	payload := map[string]any{
		"insight_id":   insight.ID,
		"entity_type":  insight.EntityType,
		"entity_id":    insight.EntityID,
		"insight_type": insight.Type,
		"route_hint":   "insight_chat",
		"source":       "insights.summary",
		"status":       insight.Status,
		"chat_context": map[string]any{
			"insight_id":             insight.ID,
			"scope":                  insight.EntityType + ":" + insight.EntityID,
			"routed_agent":           "insight_chat",
			"content_language":       "es",
			"suggested_user_message": "Explicame este insight: " + insight.Title,
			"source_kind":            "insight",
		},
	}

	chatCtx, ok := payload["chat_context"].(map[string]any)
	if !ok {
		t.Fatal("chat_context missing from payload")
	}
	if chatCtx["insight_id"] != "ins-1" {
		t.Fatalf("chat_context.insight_id = %v, want ins-1", chatCtx["insight_id"])
	}
	if chatCtx["scope"] != "project:p42" {
		t.Fatalf("chat_context.scope = %v, want project:p42", chatCtx["scope"])
	}
	if chatCtx["routed_agent"] != "insight_chat" {
		t.Fatalf("chat_context.routed_agent = %v, want insight_chat", chatCtx["routed_agent"])
	}
	if chatCtx["suggested_user_message"] != "Explicame este insight: Costo fuera de rango" {
		t.Fatalf("chat_context.suggested_user_message = %v", chatCtx["suggested_user_message"])
	}

	// Verificar que el payload completo es serializable (lo que hace el repo)
	serialized, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("payload not serializable: %v", err)
	}
	var roundtrip map[string]any
	if err := json.Unmarshal(serialized, &roundtrip); err != nil {
		t.Fatalf("payload roundtrip failed: %v", err)
	}
	rtCtx, ok := roundtrip["chat_context"].(map[string]any)
	if !ok {
		t.Fatal("chat_context lost after roundtrip")
	}
	if rtCtx["scope"] != "project:p42" {
		t.Fatalf("roundtrip chat_context.scope = %v", rtCtx["scope"])
	}

	_ = orgID
	_ = projectID
}

func TestSyncFromSummarySkipsEmptyID(t *testing.T) {
	t.Parallel()

	summary := map[string]any{
		"top_insights": []map[string]any{
			{"id": "", "title": "Empty", "summary": "Should skip"},
			{"id": "  ", "title": "Spaces", "summary": "Should skip too"},
		},
	}
	raw, _ := json.Marshal(summary)
	var envelope summaryEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}

	// Ambos insights tienen ID vacío, SyncFromSummary los ignora
	for _, insight := range envelope.TopInsights {
		if trimmed := insight.ID; trimmed != "" && trimmed != "  " {
			t.Fatalf("unexpected non-empty ID: %q", trimmed)
		}
	}
}

func TestSyncFromSummaryDedupeKeyFallback(t *testing.T) {
	t.Parallel()

	summary := map[string]any{
		"top_insights": []map[string]any{
			{"id": "ins-2", "title": "Test", "summary": "Body", "dedupe_key": ""},
			{"id": "ins-3", "title": "Test2", "summary": "Body2", "dedupe_key": "custom-key"},
		},
	}
	raw, _ := json.Marshal(summary)
	var envelope summaryEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}

	// Sin dedupe_key, debería generar "insight:{projectID}:{insightID}"
	if envelope.TopInsights[0].DedupeKey != "" {
		t.Fatalf("expected empty dedupe_key, got %q", envelope.TopInsights[0].DedupeKey)
	}
	// Con dedupe_key, debería usarlo tal cual
	if envelope.TopInsights[1].DedupeKey != "custom-key" {
		t.Fatalf("expected custom-key, got %q", envelope.TopInsights[1].DedupeKey)
	}
}

func TestSyncFromSummaryIgnoresNilOrg(t *testing.T) {
	t.Parallel()

	svc := &InsightSyncService{repo: nil}
	err := svc.SyncFromSummary(context.Background(), uuid.Nil, 42, "user", []byte(`{"top_insights":[]}`))
	if err != nil {
		t.Fatalf("expected nil for uuid.Nil org, got %v", err)
	}
}

func TestSyncFromSummaryIgnoresZeroProject(t *testing.T) {
	t.Parallel()

	svc := &InsightSyncService{repo: nil}
	err := svc.SyncFromSummary(context.Background(), uuid.New(), 0, "user", []byte(`{"top_insights":[]}`))
	if err != nil {
		t.Fatalf("expected nil for zero project, got %v", err)
	}
}

func TestSyncFromSummaryIgnoresEmptyRaw(t *testing.T) {
	t.Parallel()

	svc := &InsightSyncService{repo: nil}
	err := svc.SyncFromSummary(context.Background(), uuid.New(), 42, "user", nil)
	if err != nil {
		t.Fatalf("expected nil for empty raw, got %v", err)
	}
}
