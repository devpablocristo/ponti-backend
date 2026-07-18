package workorderdraft

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/governance"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

type stubDraftUseCases struct {
	digitalCreateCalls int
}

func (s *stubDraftUseCases) CreateWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error) {
	return 1, nil
}

func (s *stubDraftUseCases) CreateDigitalWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error) {
	s.digitalCreateCalls++
	return 7, nil
}

func (s *stubDraftUseCases) CreateDigitalWorkOrderDraftBatch(context.Context, *domain.WorkOrderDraftBatchCreate) ([]domain.WorkOrderDraftBatchCreateResultItem, error) {
	return nil, nil
}

func (s *stubDraftUseCases) PreviewDigitalWorkOrderNumber(context.Context, int64, string) (string, error) {
	return "", nil
}

func (s *stubDraftUseCases) PreviewDigitalWorkOrderDraftBatchNumber(context.Context, int64, string) (string, error) {
	return "", nil
}

func (s *stubDraftUseCases) GetWorkOrderDraftByID(context.Context, int64) (*domain.WorkOrderDraft, error) {
	return nil, nil
}

func (s *stubDraftUseCases) GetWorkOrderDraftGroupByID(context.Context, int64) (*domain.WorkOrderDraftGroup, error) {
	return nil, nil
}

func (s *stubDraftUseCases) GetWorkOrderDraftPDFData(context.Context, int64) (*pdfDocumentData, error) {
	return nil, nil
}

func (s *stubDraftUseCases) GetWorkOrderDraftGroupPDFData(context.Context, int64) (*pdfDocumentData, error) {
	return nil, nil
}

func (s *stubDraftUseCases) ListWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}

func (s *stubDraftUseCases) ListDigitalWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}

func (s *stubDraftUseCases) UpdateWorkOrderDraftByID(context.Context, *domain.WorkOrderDraft) error {
	return nil
}

func (s *stubDraftUseCases) DeleteWorkOrderDraftByID(context.Context, int64) error {
	return nil
}

func (s *stubDraftUseCases) PublishWorkOrderDraft(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *stubDraftUseCases) ListDigitalWorkOrderDraftGroups(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftGroupListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}

func (s *stubDraftUseCases) UpdateWorkOrderDraftGroupByID(context.Context, int64, *domain.WorkOrderDraftGroup) error {
	return nil
}

type draftTestEngine struct {
	router *gin.Engine
}

func (e draftTestEngine) GetRouter() *gin.Engine            { return e.router }
func (e draftTestEngine) RunServer(_ context.Context) error { return nil }

type draftTestConfig struct{}

func (draftTestConfig) APIVersion() string { return "v1" }
func (draftTestConfig) APIBaseURL() string { return "/api/v1" }

type draftTestMiddlewares struct {
	orgID uuid.UUID
	actor string
}

func (m draftTestMiddlewares) GetGlobal() []gin.HandlerFunc    { return nil }
func (m draftTestMiddlewares) GetProtected() []gin.HandlerFunc { return nil }
func (m draftTestMiddlewares) GetValidation() []gin.HandlerFunc {
	return []gin.HandlerFunc{func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxkeys.OrgID, m.orgID)
		ctx = context.WithValue(ctx, ctxkeys.Actor, m.actor)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}}
}

type stubDraftVerifier struct {
	err            error
	calls          int
	lastActionType string
}

func (s *stubDraftVerifier) VerifyApproved(_ context.Context, _ uuid.UUID, _, expectedActionType string) error {
	s.calls++
	s.lastActionType = expectedActionType
	return s.err
}

func newDraftGovernanceTestRouter(t *testing.T, actor string, ucs UseCasesPort, verifier NexusVerifierPort, verifyNexus bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewHandler(ucs, draftTestEngine{router: router}, draftTestConfig{}, draftTestMiddlewares{orgID: uuid.New(), actor: actor})
	h.SetNexusVerifier(verifier, verifyNexus)
	h.Routes()
	return router
}

func postDigitalDraft(router *gin.Engine, headers map[string]string) *httptest.ResponseRecorder {
	payload := map[string]any{
		"date":           "2026-06-01",
		"customer_id":    1,
		"project_id":     10,
		"field_id":       2,
		"lot_id":         3,
		"crop_id":        4,
		"labor_id":       5,
		"contractor":     "Contratista",
		"effective_area": 12.5,
		"is_digital":     true,
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/work-order-drafts/digital", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func TestCreateDigitalDraftBlocksAxisWithoutNexusRequestID(t *testing.T) {
	t.Parallel()
	ucs := &stubDraftUseCases{}
	verifier := &stubDraftVerifier{}
	router := newDraftGovernanceTestRouter(t, axisCompanionActor, ucs, verifier, true)

	res := postDigitalDraft(router, nil)

	if res.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d body=%s", res.Code, res.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["execution_blocked_by"] != "nexus_required" {
		t.Fatalf("expected nexus_required block, got %#v", body)
	}
	if ucs.digitalCreateCalls != 0 || verifier.calls != 0 {
		t.Fatalf("draft must not be created without header, got creates=%d verifies=%d", ucs.digitalCreateCalls, verifier.calls)
	}
}

func TestCreateDigitalDraftBlocksAxisWhenRequestNotApproved(t *testing.T) {
	t.Parallel()
	ucs := &stubDraftUseCases{}
	verifier := &stubDraftVerifier{err: &governance.NotApprovedError{Detail: "nexus request not found"}}
	router := newDraftGovernanceTestRouter(t, axisCompanionActor, ucs, verifier, true)

	res := postDigitalDraft(router, map[string]string{nexusRequestIDHeader: "req-unknown"})

	if res.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d body=%s", res.Code, res.Body.String())
	}
	if ucs.digitalCreateCalls != 0 {
		t.Fatalf("draft must not be created when unapproved, got %d", ucs.digitalCreateCalls)
	}
}

func TestCreateDigitalDraftFailsClosedWhenNexusUnreachable(t *testing.T) {
	t.Parallel()
	ucs := &stubDraftUseCases{}
	verifier := &stubDraftVerifier{err: domainerr.UpstreamError("nexus verification failed")}
	router := newDraftGovernanceTestRouter(t, axisCompanionActor, ucs, verifier, true)

	res := postDigitalDraft(router, map[string]string{nexusRequestIDHeader: "req-1"})

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 (fail closed), got %d body=%s", res.Code, res.Body.String())
	}
	if ucs.digitalCreateCalls != 0 {
		t.Fatalf("draft must not be created with nexus down, got %d", ucs.digitalCreateCalls)
	}
}

func TestCreateDigitalDraftAllowsVerifiedAxisRequest(t *testing.T) {
	t.Parallel()
	ucs := &stubDraftUseCases{}
	verifier := &stubDraftVerifier{}
	router := newDraftGovernanceTestRouter(t, axisCompanionActor, ucs, verifier, true)

	res := postDigitalDraft(router, map[string]string{nexusRequestIDHeader: "req-approved"})

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", res.Code, res.Body.String())
	}
	if ucs.digitalCreateCalls != 1 || verifier.calls != 1 {
		t.Fatalf("expected verified create, got creates=%d verifies=%d", ucs.digitalCreateCalls, verifier.calls)
	}
	if verifier.lastActionType != "ponti.workorder.draft.create" {
		t.Fatalf("expected workorder draft action type, got %q", verifier.lastActionType)
	}
}

func TestCreateDigitalDraftSkipsGateForHumanUsers(t *testing.T) {
	t.Parallel()
	ucs := &stubDraftUseCases{}
	verifier := &stubDraftVerifier{err: &governance.NotApprovedError{Detail: "must not be called"}}
	router := newDraftGovernanceTestRouter(t, "user@example.com", ucs, verifier, true)

	res := postDigitalDraft(router, nil)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201 for human users, got %d body=%s", res.Code, res.Body.String())
	}
	if verifier.calls != 0 || ucs.digitalCreateCalls != 1 {
		t.Fatalf("gate must not run for humans, got verifies=%d creates=%d", verifier.calls, ucs.digitalCreateCalls)
	}
}

func TestCreateDigitalDraftLogsOnlyWhenVerifyFlagOff(t *testing.T) {
	t.Parallel()
	ucs := &stubDraftUseCases{}
	verifier := &stubDraftVerifier{}
	router := newDraftGovernanceTestRouter(t, axisCompanionActor, ucs, verifier, false)

	res := postDigitalDraft(router, nil)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201 with flag off (log-only), got %d body=%s", res.Code, res.Body.String())
	}
	if ucs.digitalCreateCalls != 1 || verifier.calls != 0 {
		t.Fatalf("expected current behavior with flag off, got creates=%d verifies=%d", ucs.digitalCreateCalls, verifier.calls)
	}
}
