package workorderdraft

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

type workOrderDraftUseCasesStub struct {
	deletedID int64
}

func (s *workOrderDraftUseCasesStub) CreateWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error) {
	return 0, nil
}
func (s *workOrderDraftUseCasesStub) CreateDigitalWorkOrderDraft(context.Context, *domain.WorkOrderDraft) (int64, error) {
	return 0, nil
}
func (s *workOrderDraftUseCasesStub) CreateDigitalWorkOrderDraftBatch(context.Context, *domain.WorkOrderDraftBatchCreate) ([]domain.WorkOrderDraftBatchCreateResultItem, error) {
	return nil, nil
}
func (s *workOrderDraftUseCasesStub) PreviewDigitalWorkOrderNumber(context.Context, int64, string) (string, error) {
	return "", nil
}
func (s *workOrderDraftUseCasesStub) PreviewDigitalWorkOrderDraftBatchNumber(context.Context, int64, string) (string, error) {
	return "", nil
}
func (s *workOrderDraftUseCasesStub) GetWorkOrderDraftByID(context.Context, int64) (*domain.WorkOrderDraft, error) {
	return nil, nil
}
func (s *workOrderDraftUseCasesStub) GetWorkOrderDraftGroupByID(context.Context, int64) (*domain.WorkOrderDraftGroup, error) {
	return nil, nil
}
func (s *workOrderDraftUseCasesStub) GetWorkOrderDraftPDFData(context.Context, int64) (*pdfDocumentData, error) {
	return nil, nil
}
func (s *workOrderDraftUseCasesStub) GetWorkOrderDraftGroupPDFData(context.Context, int64) (*pdfDocumentData, error) {
	return nil, nil
}
func (s *workOrderDraftUseCasesStub) ListWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderDraftUseCasesStub) ListDigitalWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderDraftUseCasesStub) ListArchivedWorkOrderDrafts(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderDraftUseCasesStub) ListDigitalWorkOrderDraftGroups(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftGroupListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderDraftUseCasesStub) UpdateWorkOrderDraftByID(context.Context, *domain.WorkOrderDraft) error {
	return nil
}
func (s *workOrderDraftUseCasesStub) UpdateWorkOrderDraftGroupByID(context.Context, int64, *domain.WorkOrderDraftGroup) error {
	return nil
}
func (s *workOrderDraftUseCasesStub) DeleteWorkOrderDraftByID(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}
func (s *workOrderDraftUseCasesStub) ArchiveWorkOrderDraftByID(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}
func (s *workOrderDraftUseCasesStub) RestoreWorkOrderDraftByID(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}
func (s *workOrderDraftUseCasesStub) HardDeleteWorkOrderDraftByID(_ context.Context, id int64) error {
	s.deletedID = id
	return nil
}
func (s *workOrderDraftUseCasesStub) PublishWorkOrderDraft(context.Context, int64) (int64, error) {
	return 0, nil
}

type workOrderDraftGinEngineStub struct{ router *gin.Engine }

func (s *workOrderDraftGinEngineStub) GetRouter() *gin.Engine          { return s.router }
func (s *workOrderDraftGinEngineStub) RunServer(context.Context) error { return nil }

type workOrderDraftConfigStub struct{}

func (workOrderDraftConfigStub) APIVersion() string { return "v1" }
func (workOrderDraftConfigStub) APIBaseURL() string { return "/api/v1" }

type workOrderDraftMiddlewaresStub struct{}

func (workOrderDraftMiddlewaresStub) GetGlobal() []gin.HandlerFunc     { return nil }
func (workOrderDraftMiddlewaresStub) GetValidation() []gin.HandlerFunc { return nil }

func setupWorkOrderDraftRouter(ucs *workOrderDraftUseCasesStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(
		ucs,
		&workOrderDraftGinEngineStub{router: router},
		workOrderDraftConfigStub{},
		workOrderDraftMiddlewaresStub{},
	).Routes()
	return router
}

func TestWorkOrderDraftDeleteRouteCallsDeleteUseCase(t *testing.T) {
	ucs := &workOrderDraftUseCasesStub{}
	router := setupWorkOrderDraftRouter(ucs)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/work-order-drafts/31", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d. body=%s", rec.Code, rec.Body.String())
	}
	if ucs.deletedID != 31 {
		t.Fatalf("expected delete id 31, got %d", ucs.deletedID)
	}
}
