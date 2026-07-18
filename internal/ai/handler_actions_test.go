package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
)

type aiHandlerTestEngine struct {
	router *gin.Engine
}

func (e aiHandlerTestEngine) GetRouter() *gin.Engine { return e.router }

type aiHandlerTestConfig struct{}

func (aiHandlerTestConfig) APIVersion() string { return "v1" }
func (aiHandlerTestConfig) APIBaseURL() string { return "/api/v1" }

type aiHandlerTestMiddlewares struct {
	orgID uuid.UUID
	actor string
}

func (m aiHandlerTestMiddlewares) GetValidation() []gin.HandlerFunc {
	return []gin.HandlerFunc{func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkeys.OrgID, m.orgID)
		ctx = context.WithValue(ctx, ctxkeys.Actor, m.actor)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(ctxkeys.OrgID), m.orgID)
		c.Set(string(ctxkeys.Actor), m.actor)
		c.Next()
	}}
}

func TestPrepareInsightResolveReturnsPreviewOnlyGovernedPayload(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newAIActionTestRouter(orgID, "user-1")

	res := postJSON(router, "/api/v1/ai/actions/insight-resolve/prepare", map[string]any{
		"insight_id":      uuid.NewString(),
		"resolution_note": "Lo reviso y lo doy por resuelto.",
		"workspace": map[string]any{
			"project_id": 10,
		},
	})

	assertPreviewResponse(t, res, orgID, "ponti.insight.resolve.prepare", pontiActionTypeInsightResolve)
}

func TestPrepareWorkOrderDraftReturnsPreviewOnlyGovernedPayload(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newAIActionTestRouter(orgID, "user-2")

	res := postJSON(router, "/api/v1/ai/actions/workorder-draft/prepare", map[string]any{
		"project_id":     10,
		"field_id":       20,
		"campaign_id":    30,
		"work_type":      "siembra",
		"scheduled_date": "2026-07-01",
		"notes":          "Preparar borrador, no publicar.",
	})

	body := assertPreviewResponse(t, res, orgID, "ponti.workorder.draft.prepare", pontiActionTypeWorkOrderDraftCreate)
	evidence := body["evidence"].(map[string]any)
	workspace := evidence["workspace"].(map[string]any)
	if workspace["project_id"].(float64) != 10 {
		t.Fatalf("expected project_id in workspace, got %#v", workspace)
	}
}

func TestPrepareStockAdjustmentReturnsPreviewOnlyGovernedPayload(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newAIActionTestRouter(orgID, "user-3")

	res := postJSON(router, "/api/v1/ai/actions/stock-adjustment/prepare", map[string]any{
		"project_id":     10,
		"supply_id":      5,
		"quantity_delta": -12.5,
		"reason":         "Correccion propuesta por diferencia detectada.",
	})

	assertPreviewResponse(t, res, orgID, "ponti.stock_adjustment.prepare", pontiActionTypeStockAdjust)
}

func TestPrepareStockAdjustmentRejectsZeroDelta(t *testing.T) {
	t.Parallel()
	router := newAIActionTestRouter(uuid.New(), "user-4")

	res := postJSON(router, "/api/v1/ai/actions/stock-adjustment/prepare", map[string]any{
		"project_id":     10,
		"supply_id":      5,
		"quantity_delta": 0,
		"reason":         "Sin cambio.",
	})

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", res.Code, res.Body.String())
	}
}

func TestDraftInsightResolutionReturnsDraftContract(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newAIActionTestRouter(orgID, "user-6")

	res := postJSONWithHeaders(router, "/api/v1/ai/actions/insight-resolution/draft", map[string]any{
		"insight_id":      uuid.NewString(),
		"resolution_note": "Dejar trazado para revision.",
		"workspace": map[string]any{
			"project_id": 10,
		},
	}, map[string]string{"X-Nexus-Request-ID": uuid.NewString()})

	assertDraftExecutionResponse(t, res, orgID, "ponti.insight_resolution.draft", false)
}

func TestDraftStockCountReturnsDraftContract(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	router := newAIActionTestRouter(orgID, "user-7")
	nexusRequestID := uuid.NewString()

	res := postJSONWithHeaders(router, "/api/v1/ai/actions/stock-count/draft", map[string]any{
		"project_id":       10,
		"supply_id":        5,
		"real_stock_units": 12.5,
		"reason":           "Conteo preparado por diferencia detectada.",
	}, map[string]string{"X-Nexus-Request-ID": nexusRequestID})

	body := assertDraftExecutionResponse(t, res, orgID, "ponti.stock_count.draft", false)
	if body["nexus_request_id"] != nexusRequestID {
		t.Fatalf("expected nexus request id propagated, got %#v", body)
	}
}

func TestPrepareWorkOrderDraftRejectsInvalidDate(t *testing.T) {
	t.Parallel()
	router := newAIActionTestRouter(uuid.New(), "user-5")

	res := postJSON(router, "/api/v1/ai/actions/workorder-draft/prepare", map[string]any{
		"project_id":     10,
		"work_type":      "siembra",
		"scheduled_date": "01/07/2026",
	})

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", res.Code, res.Body.String())
	}
}

func newAIActionTestRouter(orgID uuid.UUID, actor string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewHandler(nil, aiHandlerTestEngine{router: router}, aiHandlerTestConfig{}, aiHandlerTestMiddlewares{orgID: orgID, actor: actor})
	h.Routes()
	return router
}

func postJSON(router *gin.Engine, path string, payload map[string]any) *httptest.ResponseRecorder {
	return postJSONWithHeaders(router, path, payload, nil)
}

func postJSONWithHeaders(router *gin.Engine, path string, payload map[string]any, headers map[string]string) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func assertPreviewResponse(t *testing.T, res *httptest.ResponseRecorder, orgID uuid.UUID, action, actionType string) map[string]any {
	t.Helper()
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != draftActionStatusPreview {
		t.Fatalf("unexpected status: %#v", body)
	}
	if body["action"] != action {
		t.Fatalf("unexpected action: %#v", body)
	}
	if body["approval_required"] != true || body["nexus_action_type"] != actionType {
		t.Fatalf("governance missing: %#v", body)
	}
	if body["preview_only"] != true || body["write_performed"] != false || body["execution_allowed"] != false {
		t.Fatalf("preview-only guard missing: %#v", body)
	}
	evidence := body["evidence"].(map[string]any)
	if evidence["tenant_scope"] != orgID.String() {
		t.Fatalf("tenant evidence mismatch: %#v", evidence)
	}
	if evidence["approval_required"] != true {
		t.Fatalf("approval evidence missing: %#v", evidence)
	}
	return body
}

func assertDraftExecutionResponse(t *testing.T, res *httptest.ResponseRecorder, orgID uuid.UUID, action string, writePerformed bool) map[string]any {
	t.Helper()
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["action"] != action {
		t.Fatalf("unexpected action: %#v", body)
	}
	if body["write_performed"] != writePerformed {
		t.Fatalf("unexpected write_performed: %#v", body)
	}
	if body["draft_id"] == "" || body["execution_status"] == "" || body["audit_ref"] == "" {
		t.Fatalf("draft execution contract missing: %#v", body)
	}
	evidence := body["evidence"].(map[string]any)
	if evidence["tenant_scope"] != orgID.String() {
		t.Fatalf("tenant evidence mismatch: %#v", evidence)
	}
	if evidence["approval_required"] != true {
		t.Fatalf("approval evidence missing: %#v", evidence)
	}
	return body
}
