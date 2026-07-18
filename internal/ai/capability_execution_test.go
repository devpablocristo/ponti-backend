package ai

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestCapabilityExecutionRejectsInvalidSchema(t *testing.T) {
	t.Parallel()
	router := newAIActionTestRouter(uuid.New(), axisCompanionActor)

	res := postJSON(router, "/api/v1/capability-executions", map[string]any{
		"schema_version": "capability_execution.v0",
		"operation":      "ponti.workorder.draft.prepare",
		"payload":        map[string]any{},
	})

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestCapabilityExecutionRejectsUnknownOperation(t *testing.T) {
	t.Parallel()
	router := newAIActionTestRouter(uuid.New(), axisCompanionActor)

	res := postJSON(router, "/api/v1/capability-executions", map[string]any{
		"schema_version": capabilityExecutionSchemaVersion,
		"operation":      "ponti.unknown",
		"payload":        map[string]any{},
	})

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestCapabilityExecutionDispatchesExistingAction(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newAIActionTestRouter(orgID, axisCompanionActor)

	res := postJSON(router, "/api/v1/capability-executions", map[string]any{
		"schema_version": capabilityExecutionSchemaVersion,
		"operation":      "ponti.workorder.draft.prepare",
		"executor_ref":   "ponti-backend.actions.workorder.draft.prepare",
		"payload": map[string]any{
			"project_id":     10,
			"field_id":       20,
			"campaign_id":    30,
			"work_type":      "siembra",
			"scheduled_date": "2026-07-01",
			"workspace": map[string]any{
				"project_id":  10,
				"campaign_id": 30,
			},
		},
		"actor": map[string]any{
			"actor_id":        "axis-runner",
			"actor_type":      "agent",
			"product_surface": "ponti",
		},
		"org_id": "axis-local-org",
	})

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "success" {
		t.Fatalf("unexpected envelope status: %#v", body)
	}
	result := body["result"].(map[string]any)
	if result["status"] != draftActionStatusPreview {
		t.Fatalf("expected preview result, got %#v", result)
	}
	evidence := body["evidence"].(map[string]any)
	if evidence["tenant_scope"] != orgID.String() {
		t.Fatalf("tenant evidence mismatch: %#v", evidence)
	}
	if evidence["internal_route"] != "POST /api/v1/ai/actions/workorder-draft/prepare" {
		t.Fatalf("unexpected route evidence: %#v", evidence)
	}
}

func TestCapabilityExecutionCoversPublishedTools(t *testing.T) {
	t.Parallel()
	h := &Handler{acf: aiHandlerTestConfig{}}

	for _, manifest := range pontiCapabilities() {
		for _, tool := range manifest.Tools {
			tool := tool
			t.Run(tool.Name, func(t *testing.T) {
				t.Parallel()
				_, _, err := h.capabilityExecutionTarget(capabilityExecutionRequest{
					SchemaVersion: capabilityExecutionSchemaVersion,
					Operation:     tool.Name,
					Payload:       capabilityExecutionSamplePayload(t, tool.Name),
				})
				if err != nil {
					t.Fatalf("published capability is not executable: %v", err)
				}
			})
		}
	}
}

func capabilityExecutionSamplePayload(t *testing.T, operation string) json.RawMessage {
	t.Helper()
	payload := map[string]any{
		"workspace": map[string]any{
			"project_id":  10,
			"campaign_id": 30,
		},
	}
	switch operation {
	case "ponti.insights.explain", "ponti.insight.resolve.prepare", "ponti.insight_resolution.draft":
		payload["insight_id"] = uuid.NewString()
		payload["resolution_note"] = "Revision propuesta por Axis."
	case "ponti.stock.summary":
		payload["project_id"] = 10
	case "ponti.workorder.draft.prepare", "ponti.workorder_draft.create":
		payload["project_id"] = 10
		payload["field_id"] = 20
		payload["campaign_id"] = 30
		payload["work_type"] = "siembra"
		payload["scheduled_date"] = "2026-07-01"
	case "ponti.stock_adjustment.prepare":
		payload["project_id"] = 10
		payload["supply_id"] = 5
		payload["quantity_delta"] = -3.5
		payload["reason"] = "Ajuste propuesto por diferencia."
	case "ponti.stock_count.draft":
		payload["project_id"] = 10
		payload["supply_id"] = 5
		payload["real_stock_units"] = 12.5
		payload["reason"] = "Conteo preparado."
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
		return nil
	}
	return raw
}
