package workorder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type workOrderHandlerUseCasesStub struct {
	updateInvestorPaymentStatusFn func(context.Context, int64, int64, string) error
	actionCall                    string

	updateWorkOrderIDCalls []struct {
		workOrderID int64
		investorID  int64
		status      string
	}
}

func (s *workOrderHandlerUseCasesStub) CreateWorkOrder(context.Context, *domain.WorkOrder) (int64, error) {
	return 0, nil
}
func (s *workOrderHandlerUseCasesStub) GetWorkOrderByID(context.Context, int64) (*domain.WorkOrder, error) {
	return nil, nil
}
func (s *workOrderHandlerUseCasesStub) DuplicateWorkOrder(context.Context, string) (string, error) {
	return "", nil
}
func (s *workOrderHandlerUseCasesStub) UpdateWorkOrderByID(context.Context, *domain.WorkOrder) error {
	return nil
}
func (s *workOrderHandlerUseCasesStub) UpdateInvestorPaymentStatus(
	ctx context.Context,
	workOrderID int64,
	investorID int64,
	status string,
) error {
	s.updateWorkOrderIDCalls = append(s.updateWorkOrderIDCalls, struct {
		workOrderID int64
		investorID  int64
		status      string
	}{
		workOrderID: workOrderID,
		investorID:  investorID,
		status:      status,
	})
	if s.updateInvestorPaymentStatusFn != nil {
		return s.updateInvestorPaymentStatusFn(ctx, workOrderID, investorID, status)
	}
	return nil
}
func (s *workOrderHandlerUseCasesStub) DeleteWorkOrderByID(context.Context, int64) error {
	s.actionCall = "delete"
	return nil
}
func (s *workOrderHandlerUseCasesStub) HardDeleteWorkOrder(context.Context, int64) error {
	s.actionCall = "hard"
	return nil
}
func (s *workOrderHandlerUseCasesStub) ArchiveWorkOrder(context.Context, int64) error {
	s.actionCall = "archive"
	return nil
}
func (s *workOrderHandlerUseCasesStub) RestoreWorkOrder(context.Context, int64) error {
	s.actionCall = "restore"
	return nil
}
func (s *workOrderHandlerUseCasesStub) ListArchivedWorkOrders(context.Context, int, int) ([]domain.WorkOrderListElement, int64, error) {
	return nil, 0, nil
}
func (s *workOrderHandlerUseCasesStub) ListWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderHandlerUseCasesStub) ListWorkOrderFilterRows(context.Context, domain.WorkOrderFilter) ([]domain.WorkOrderListElement, error) {
	return nil, nil
}
func (s *workOrderHandlerUseCasesStub) GetMetrics(context.Context, domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error) {
	return nil, nil
}
func (s *workOrderHandlerUseCasesStub) ExportWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]byte, error) {
	return nil, nil
}

type workOrderGinEngineStub struct{ router *gin.Engine }

func (s *workOrderGinEngineStub) GetRouter() *gin.Engine          { return s.router }
func (s *workOrderGinEngineStub) RunServer(context.Context) error { return nil }

type workOrderConfigStub struct{}

func (workOrderConfigStub) APIVersion() string { return "v1" }
func (workOrderConfigStub) APIBaseURL() string { return "/api/v1" }

type workOrderMiddlewaresStub struct{}

func (workOrderMiddlewaresStub) GetGlobal() []gin.HandlerFunc     { return nil }
func (workOrderMiddlewaresStub) GetValidation() []gin.HandlerFunc { return nil }

func setupWorkOrderRouter(ucs *workOrderHandlerUseCasesStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(
		ucs,
		&workOrderGinEngineStub{router: router},
		workOrderConfigStub{},
		workOrderMiddlewaresStub{},
	).Routes()
	return router
}

func TestWorkOrderActionRoutesCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantAction string
	}{
		{
			name:       "archive route calls archive usecase",
			method:     http.MethodPost,
			path:       "/api/v1/work-orders/15/archive",
			wantAction: "archive",
		},
		{
			name:       "restore route calls restore usecase",
			method:     http.MethodPost,
			path:       "/api/v1/work-orders/15/restore",
			wantAction: "restore",
		},
		{
			name:       "explicit hard delete route calls hard-delete usecase",
			method:     http.MethodDelete,
			path:       "/api/v1/work-orders/15/hard",
			wantAction: "hard",
		},
		{
			name:       "legacy delete route calls legacy hard-delete alias",
			method:     http.MethodDelete,
			path:       "/api/v1/work-orders/15",
			wantAction: "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &workOrderHandlerUseCasesStub{}
			router := setupWorkOrderRouter(stub)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("expected status 204, got %d. body=%s", rec.Code, rec.Body.String())
			}
			if stub.actionCall != tt.wantAction {
				t.Fatalf("expected action %q, got %q", tt.wantAction, stub.actionCall)
			}
		})
	}
}

func TestHandler_UpdateInvestorPaymentStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &workOrderHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/work-orders/15/investors/9/payment-status",
		strings.NewReader(`{"payment_status":"Pagada"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "work_order_id", Value: "15"},
		{Key: "investor_id", Value: "9"},
	}

	h.UpdateInvestorPaymentStatus(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, ctx.Writer.Status())
	}
	if len(stub.updateWorkOrderIDCalls) != 1 {
		t.Fatalf("expected one use case call, got %d", len(stub.updateWorkOrderIDCalls))
	}
	call := stub.updateWorkOrderIDCalls[0]
	if call.workOrderID != 15 || call.investorID != 9 || call.status != "Pagada" {
		t.Fatalf("unexpected call payload: %+v", call)
	}
}
