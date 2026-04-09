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
	return nil
}
func (s *workOrderHandlerUseCasesStub) ArchiveWorkOrder(context.Context, int64) error {
	return nil
}
func (s *workOrderHandlerUseCasesStub) RestoreWorkOrder(context.Context, int64) error {
	return nil
}
func (s *workOrderHandlerUseCasesStub) ListWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *workOrderHandlerUseCasesStub) GetMetrics(context.Context, domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error) {
	return nil, nil
}
func (s *workOrderHandlerUseCasesStub) ExportWorkOrders(context.Context, domain.WorkOrderFilter, types.Input) ([]byte, error) {
	return nil, nil
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
