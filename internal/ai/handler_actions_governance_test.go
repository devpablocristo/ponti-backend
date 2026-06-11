package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/devpablocristo/ponti-backend/internal/governance"
)

type stubActionVerifier struct {
	err             error
	calls           int
	lastActionType  string
	lastNexusReqID  string
	lastTenant      uuid.UUID
	lastTenantValid bool
}

func (s *stubActionVerifier) VerifyApproved(_ context.Context, tenantID uuid.UUID, nexusRequestID, expectedActionType string) error {
	s.calls++
	s.lastTenant = tenantID
	s.lastTenantValid = tenantID != uuid.Nil
	s.lastNexusReqID = nexusRequestID
	s.lastActionType = expectedActionType
	return s.err
}

type stubInsightResolver struct {
	calls      int
	lastTenant uuid.UUID
	lastID     string
	lastActor  string
	err        error
}

func (s *stubInsightResolver) ResolveManual(_ context.Context, tenantID uuid.UUID, candidateID, actor string) error {
	s.calls++
	s.lastTenant = tenantID
	s.lastID = candidateID
	s.lastActor = actor
	return s.err
}

func newActionDraftTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&actionDraftModel{}); err != nil {
		t.Fatalf("auto migrate action drafts: %v", err)
	}
	return db
}

func newGovernedAIActionTestRouter(t *testing.T, orgID uuid.UUID, actor string, executor *ActionExecutor, verifier ActionVerifierPort, cfg GovernedActionsConfig) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewHandler(nil, aiHandlerTestEngine{router: router}, aiHandlerTestConfig{}, aiHandlerTestMiddlewares{orgID: orgID, actor: actor})
	h.SetGovernedActions(executor, verifier, cfg)
	h.Routes()
	return router
}

func TestDraftEndpointsRequireNexusHeaderWhenVerifyEnabledForAxis(t *testing.T) {
	t.Parallel()
	verifier := &stubActionVerifier{}
	router := newGovernedAIActionTestRouter(t, uuid.New(), axisCompanionActor, nil, verifier, GovernedActionsConfig{VerifyNexus: true})

	res := postJSON(router, "/api/v1/ai/actions/insight-resolution/draft", map[string]any{
		"insight_id": uuid.NewString(),
	})

	if res.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["execution_blocked_by"] != executionBlockedByNexus {
		t.Fatalf("expected execution_blocked_by=nexus_required, got %#v", body)
	}
	if body["detail"] == "" || body["detail"] == nil {
		t.Fatalf("expected blocking detail, got %#v", body)
	}
	if verifier.calls != 0 {
		t.Fatalf("verifier must not be called without header, got %d calls", verifier.calls)
	}
}

func TestPrepareEndpointReturns412WhenNexusRequestNotApproved(t *testing.T) {
	t.Parallel()
	verifier := &stubActionVerifier{err: &governance.NotApprovedError{Detail: `nexus request status "pending_approval" is not approved`}}
	router := newGovernedAIActionTestRouter(t, uuid.New(), axisCompanionActor, nil, verifier, GovernedActionsConfig{VerifyNexus: true})

	res := postJSONWithHeaders(router, "/api/v1/ai/actions/stock-adjustment/prepare", map[string]any{
		"project_id":     10,
		"supply_id":      5,
		"quantity_delta": -2.5,
		"reason":         "Ajuste propuesto.",
	}, map[string]string{nexusRequestIDHeader: "req-pending"})

	if res.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d body=%s", res.Code, res.Body.String())
	}
	if verifier.lastActionType != pontiActionTypeStockAdjust {
		t.Fatalf("expected per-tool action type, got %q", verifier.lastActionType)
	}
}

func TestDraftEndpointFailsClosedWhenNexusUnreachable(t *testing.T) {
	t.Parallel()
	verifier := &stubActionVerifier{err: domainerr.UpstreamError("nexus verification failed")}
	router := newGovernedAIActionTestRouter(t, uuid.New(), axisCompanionActor, nil, verifier, GovernedActionsConfig{VerifyNexus: true})

	res := postJSONWithHeaders(router, "/api/v1/ai/actions/stock-count/draft", map[string]any{
		"project_id":       10,
		"supply_id":        5,
		"real_stock_units": 7.5,
		"reason":           "Conteo de campo.",
	}, map[string]string{nexusRequestIDHeader: "req-1"})

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 (fail closed), got %d body=%s", res.Code, res.Body.String())
	}
}

func TestDraftEndpointsSkipEnforcementForNonAxisActors(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	verifier := &stubActionVerifier{err: &governance.NotApprovedError{Detail: "must not be called"}}
	router := newGovernedAIActionTestRouter(t, orgID, "user-9", nil, verifier, GovernedActionsConfig{VerifyNexus: true})

	res := postJSON(router, "/api/v1/ai/actions/insight-resolution/draft", map[string]any{
		"insight_id": uuid.NewString(),
	})

	assertDraftExecutionResponse(t, res, orgID, "ponti.insight_resolution.draft", false)
	if verifier.calls != 0 {
		t.Fatalf("verifier must not run for non-axis actors, got %d calls", verifier.calls)
	}
}

func TestDraftInsightResolutionAppliesWriteWhenVerifiedAndEnabled(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	db := newActionDraftTestDB(t)
	resolver := &stubInsightResolver{}
	verifier := &stubActionVerifier{}
	executor := NewActionExecutor(NewActionDraftRepository(db), resolver, nil, verifier, ActionExecutorConfig{GovernedWritesEnabled: true})
	router := newGovernedAIActionTestRouter(t, orgID, axisCompanionActor, executor, verifier, GovernedActionsConfig{VerifyNexus: true})
	insightID := uuid.NewString()

	res := postJSONWithHeaders(router, "/api/v1/ai/actions/insight-resolution/draft", map[string]any{
		"insight_id":      insightID,
		"resolution_note": "Aplicar resolución aprobada.",
	}, map[string]string{nexusRequestIDHeader: "req-approved"})

	body := assertDraftExecutionResponse(t, res, orgID, "ponti.insight_resolution.draft", true)
	if body["preview_only"] != false || body["execution_status"] != "applied" {
		t.Fatalf("expected applied execution, got %#v", body)
	}
	if resolver.calls != 1 || resolver.lastID != insightID || resolver.lastTenant != orgID {
		t.Fatalf("expected insight resolution applied, got %+v", resolver)
	}
	var draft actionDraftModel
	if err := db.First(&draft, "tenant_id = ? AND nexus_request_id = ?", orgID, "req-approved").Error; err != nil {
		t.Fatalf("expected persisted draft row: %v", err)
	}
	if draft.Status != actionDraftStatusApplied || draft.AppliedAt == nil {
		t.Fatalf("expected applied draft row, got %+v", draft)
	}
}

func TestDraftStockCountStaysStagedWhenWritesDisabled(t *testing.T) {
	t.Parallel()
	orgID := uuid.New()
	db := newActionDraftTestDB(t)
	verifier := &stubActionVerifier{}
	executor := NewActionExecutor(NewActionDraftRepository(db), nil, nil, verifier, ActionExecutorConfig{GovernedWritesEnabled: false})
	router := newGovernedAIActionTestRouter(t, orgID, axisCompanionActor, executor, verifier, GovernedActionsConfig{VerifyNexus: true})

	res := postJSONWithHeaders(router, "/api/v1/ai/actions/stock-count/draft", map[string]any{
		"project_id":       10,
		"supply_id":        5,
		"real_stock_units": 12.5,
		"reason":           "Conteo preparado.",
	}, map[string]string{nexusRequestIDHeader: "req-approved"})

	body := assertDraftExecutionResponse(t, res, orgID, "ponti.stock_count.draft", false)
	if body["preview_only"] != true || body["execution_status"] != "draft_staged" {
		t.Fatalf("expected staged preview, got %#v", body)
	}
	var draft actionDraftModel
	if err := db.First(&draft, "tenant_id = ? AND nexus_request_id = ?", orgID, "req-approved").Error; err != nil {
		t.Fatalf("expected persisted draft row: %v", err)
	}
	if draft.Status != actionDraftStatusStaged {
		t.Fatalf("expected staged draft row, got %+v", draft)
	}
}
