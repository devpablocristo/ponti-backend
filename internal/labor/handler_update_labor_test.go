package labor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/alphacodinggroup/ponti-backend/internal/labor/usecases/domain"
	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type laborHandlerUseCasesStub struct {
	getLaborFn      func(ctx context.Context, id int64) (*domain.Labor, error)
	updateLaborFn   func(ctx context.Context, labor *domain.Labor) error
	getLaborCalls   []int64
	updateLaborCall []domain.Labor
}

func (s *laborHandlerUseCasesStub) CreateLabor(context.Context, *domain.Labor) (int64, error) {
	return 0, nil
}
func (s *laborHandlerUseCasesStub) GetLabor(ctx context.Context, id int64) (*domain.Labor, error) {
	s.getLaborCalls = append(s.getLaborCalls, id)
	if s.getLaborFn != nil {
		return s.getLaborFn(ctx, id)
	}
	return &domain.Labor{ID: id}, nil
}
func (s *laborHandlerUseCasesStub) ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error) {
	return nil, 0, nil
}
func (s *laborHandlerUseCasesStub) DeleteLabor(context.Context, int64) error { return nil }
func (s *laborHandlerUseCasesStub) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	s.updateLaborCall = append(s.updateLaborCall, *labor)
	if s.updateLaborFn != nil {
		return s.updateLaborFn(ctx, labor)
	}
	return nil
}
func (s *laborHandlerUseCasesStub) CountWorkOrdersByLaborID(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s *laborHandlerUseCasesStub) ListLaborCategoriesByTypeID(context.Context, int64) ([]domain.LaborCategory, error) {
	return nil, nil
}
func (s *laborHandlerUseCasesStub) ListLaborByWorkOrder(context.Context, int64) ([]domain.LaborRawItem, error) {
	return nil, nil
}
func (s *laborHandlerUseCasesStub) ListGroupLaborByWorkOrder(context.Context, types.Input, int64, int64) ([]domain.LaborListItem, types.PageInfo, error) {
	return nil, types.PageInfo{}, nil
}
func (s *laborHandlerUseCasesStub) ExportGroupLaborXLSX(context.Context, types.Input, int64, int64) ([]byte, error) {
	return nil, nil
}
func (s *laborHandlerUseCasesStub) ExportAllGroupLabors(context.Context, int64) ([]byte, error) {
	return nil, nil
}
func (s *laborHandlerUseCasesStub) GetMetrics(context.Context, domain.LaborFilter) (*domain.LaborMetrics, error) {
	return nil, nil
}

func newLaborHandlerJSONContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// UpdateLabor requiere user_id en contexto como string
	req = req.WithContext(context.WithValue(req.Context(), pkgmwr.ContextUserIDKey, "123"))

	ctx.Request = req
	return ctx, rec
}

func TestHandler_UpdateLabor_OmittedIsPartialPrice_PreservesStoredValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &laborHandlerUseCasesStub{
		getLaborFn: func(_ context.Context, id int64) (*domain.Labor, error) {
			return &domain.Labor{ID: id, IsPartialPrice: true}, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, _ := newLaborHandlerJSONContext(http.MethodPut, "/api/v1/projects/10/labors/42", `{
		"name": "Siembra",
		"contractor_name": "Contratista A",
		"price": "125.00",
		"category_id": 3
	}`)
	ctx.Params = gin.Params{
		{Key: "project_id", Value: "10"},
		{Key: "labor_id", Value: "42"},
	}

	h.UpdateLabor(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, ctx.Writer.Status())
	}
	if len(stub.getLaborCalls) != 1 || stub.getLaborCalls[0] != 42 {
		t.Fatalf("expected GetLabor to be called with id 42, got %#v", stub.getLaborCalls)
	}
	if len(stub.updateLaborCall) != 1 {
		t.Fatalf("expected one UpdateLabor call, got %d", len(stub.updateLaborCall))
	}
	if !stub.updateLaborCall[0].IsPartialPrice {
		t.Fatalf("expected IsPartialPrice=true to be preserved")
	}
}

func TestHandler_UpdateLabor_ExplicitIsPartialPrice_DoesNotFetchCurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &laborHandlerUseCasesStub{
		getLaborFn: func(_ context.Context, _ int64) (*domain.Labor, error) {
			t.Fatalf("GetLabor should not be called when is_partial_price is explicit")
			return nil, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, _ := newLaborHandlerJSONContext(http.MethodPut, "/api/v1/projects/10/labors/42", `{
		"name": "Siembra",
		"contractor_name": "Contratista A",
		"price": "125.00",
		"category_id": 3,
		"is_partial_price": false
	}`)
	ctx.Params = gin.Params{
		{Key: "project_id", Value: "10"},
		{Key: "labor_id", Value: "42"},
	}

	h.UpdateLabor(ctx)

	if ctx.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, ctx.Writer.Status())
	}
	if len(stub.getLaborCalls) != 0 {
		t.Fatalf("expected no GetLabor call, got %#v", stub.getLaborCalls)
	}
	if len(stub.updateLaborCall) != 1 {
		t.Fatalf("expected one UpdateLabor call, got %d", len(stub.updateLaborCall))
	}
	if stub.updateLaborCall[0].IsPartialPrice {
		t.Fatalf("expected IsPartialPrice=false from explicit payload")
	}
}
