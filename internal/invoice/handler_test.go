package invoice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
)

type invoiceHandlerUseCasesStub struct {
	getCalls    []invoiceTarget
	deleteCalls []invoiceTarget
	listCalls   []struct {
		projectID int64
		page      int
		perPage   int
	}
}

func (s *invoiceHandlerUseCasesStub) GetInvoiceByWorkOrder(_ context.Context, workOrderID int64, investorID int64) (*domain.Invoice, error) {
	s.getCalls = append(s.getCalls, invoiceTarget{workOrderID: workOrderID, investorID: investorID})
	return &domain.Invoice{ID: workOrderID, WorkOrderID: workOrderID, InvestorID: investorID}, nil
}

func (s *invoiceHandlerUseCasesStub) CreateInvoice(context.Context, *domain.Invoice) (int64, error) {
	return 0, nil
}

func (s *invoiceHandlerUseCasesStub) UpdateInvoice(context.Context, *domain.Invoice) error {
	return nil
}

func (s *invoiceHandlerUseCasesStub) DeleteInvoice(_ context.Context, workOrderID int64, investorID int64) error {
	s.deleteCalls = append(s.deleteCalls, invoiceTarget{workOrderID: workOrderID, investorID: investorID})
	return nil
}

func (s *invoiceHandlerUseCasesStub) ListInvoices(_ context.Context, projectID int64, page int, perPage int) ([]domain.Invoice, int64, error) {
	s.listCalls = append(s.listCalls, struct {
		projectID int64
		page      int
		perPage   int
	}{projectID: projectID, page: page, perPage: perPage})
	return []domain.Invoice{{ID: 1}}, 1, nil
}

func newInvoiceHandlerContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(""))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_GetInvoiceByWorkOrder_ParsesWorkOrderAndInvestor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &invoiceHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newInvoiceHandlerContext(http.MethodGet, "/api/v1/invoices/42?investor_id=9")
	ctx.Params = gin.Params{{Key: "work_order_id", Value: "42"}}

	h.GetInvoiceByWorkOrder(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.getCalls) != 1 || stub.getCalls[0].workOrderID != 42 || stub.getCalls[0].investorID != 9 {
		t.Fatalf("expected get call with work order 42 and investor 9, got %#v", stub.getCalls)
	}
}

func TestHandler_DeleteInvoice_ParsesWorkOrderAndInvestor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &invoiceHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newInvoiceHandlerContext(http.MethodDelete, "/api/v1/invoices/42?investor_id=9")
	ctx.Params = gin.Params{{Key: "work_order_id", Value: "42"}}

	h.DeleteInvoice(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, ctx.Writer.Status())
	}
	if len(stub.deleteCalls) != 1 || stub.deleteCalls[0].workOrderID != 42 || stub.deleteCalls[0].investorID != 9 {
		t.Fatalf("expected delete call with work order 42 and investor 9, got %#v", stub.deleteCalls)
	}
}

func TestHandler_ListInvoices_ParsesProjectAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &invoiceHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newInvoiceHandlerContext(http.MethodGet, "/api/v1/invoices?project_id=10&page=2&per_page=25")

	h.ListInvoices(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.listCalls) != 1 {
		t.Fatalf("expected one list call, got %#v", stub.listCalls)
	}
	if got := stub.listCalls[0]; got.projectID != 10 || got.page != 2 || got.perPage != 25 {
		t.Fatalf("expected project 10 page 2 per_page 25, got %#v", got)
	}
}
