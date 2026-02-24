package supply

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	providerdomain "github.com/alphacodinggroup/ponti-backend/internal/provider/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
)

type handlerUseCasesStub struct {
	getSupplyFn         func(ctx context.Context, id int64) (*domain.Supply, error)
	getSuppliesByIDsFn  func(ctx context.Context, ids []int64) (map[int64]domain.Supply, error)
	updateSupplyFn      func(ctx context.Context, s *domain.Supply) error
	updateSuppliesFn    func(ctx context.Context, supplies []domain.Supply) error
	getSupplyCalls      []int64
	getSuppliesIDsCalls [][]int64
	updateSupplyCalls   []domain.Supply
	updateBulkCalls     [][]domain.Supply
}

func (s *handlerUseCasesStub) CreateSupply(context.Context, *domain.Supply) (int64, error) {
	return 0, nil
}

func (s *handlerUseCasesStub) CreateSuppliesBulk(context.Context, []domain.Supply) error {
	return nil
}

func (s *handlerUseCasesStub) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	s.getSupplyCalls = append(s.getSupplyCalls, id)
	if s.getSupplyFn != nil {
		return s.getSupplyFn(ctx, id)
	}
	return nil, nil
}

func (s *handlerUseCasesStub) GetSuppliesByIDs(ctx context.Context, ids []int64) (map[int64]domain.Supply, error) {
	idsCopy := append([]int64(nil), ids...)
	s.getSuppliesIDsCalls = append(s.getSuppliesIDsCalls, idsCopy)
	if s.getSuppliesByIDsFn != nil {
		return s.getSuppliesByIDsFn(ctx, ids)
	}
	return map[int64]domain.Supply{}, nil
}

func (s *handlerUseCasesStub) UpdateSupply(ctx context.Context, supply *domain.Supply) error {
	s.updateSupplyCalls = append(s.updateSupplyCalls, *supply)
	if s.updateSupplyFn != nil {
		return s.updateSupplyFn(ctx, supply)
	}
	return nil
}

func (s *handlerUseCasesStub) DeleteSupply(context.Context, int64) error {
	return nil
}

func (s *handlerUseCasesStub) CountWorkOrdersBySupplyID(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *handlerUseCasesStub) ListSuppliesPaginated(context.Context, domain.SupplyFilter, int, int, string) ([]domain.Supply, int64, error) {
	return nil, 0, nil
}

func (s *handlerUseCasesStub) UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	bulkCopy := append([]domain.Supply(nil), supplies...)
	s.updateBulkCalls = append(s.updateBulkCalls, bulkCopy)
	if s.updateSuppliesFn != nil {
		return s.updateSuppliesFn(ctx, supplies)
	}
	return nil
}

func (s *handlerUseCasesStub) ExportTableSupplies(context.Context, domain.SupplyFilter) ([]byte, error) {
	return nil, nil
}

func (s *handlerUseCasesStub) GetEntriesSupplyMovementsByProjectID(context.Context, int64) ([]*domain.SupplyMovement, error) {
	return nil, nil
}

func (s *handlerUseCasesStub) CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error) {
	return 0, nil
}

func (s *handlerUseCasesStub) ValidateSupplyMovement(context.Context, *domain.SupplyMovement) error {
	return nil
}

func (s *handlerUseCasesStub) CreateSupplyMovementsStrict(context.Context, []*domain.SupplyMovement) ([]int64, error) {
	return nil, nil
}

func (s *handlerUseCasesStub) GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error) {
	return nil, nil
}

func (s *handlerUseCasesStub) UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error {
	return nil
}

func (s *handlerUseCasesStub) GetProviders(context.Context) ([]providerdomain.Provider, error) {
	return nil, nil
}

func (s *handlerUseCasesStub) ExportSupplyMovementsByProjectID(context.Context, int64) ([]byte, error) {
	return nil, nil
}

func (s *handlerUseCasesStub) DeleteSupplyMovement(context.Context, int64, int64) error {
	return nil
}

func newHandlerJSONContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, rec
}

func TestHandler_UpdateSupply_OmittedIsPartialPrice_PreservesStoredValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		getSupplyFn: func(_ context.Context, id int64) (*domain.Supply, error) {
			return &domain.Supply{ID: id, IsPartialPrice: true}, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, _ := newHandlerJSONContext(http.MethodPut, "/api/v1/supplies/42", `{
		"project_id": 10,
		"name": "Urea",
		"price": "125.00",
		"unit_id": 2,
		"category_id": 5,
		"type_id": 3
	}`)
	ctx.Params = gin.Params{{Key: "supply_id", Value: "42"}}

	h.UpdateSupply(ctx)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
	assert.Equal(t, []int64{42}, stub.getSupplyCalls)
	if assert.Len(t, stub.updateSupplyCalls, 1) {
		assert.Equal(t, int64(42), stub.updateSupplyCalls[0].ID)
		assert.True(t, stub.updateSupplyCalls[0].IsPartialPrice)
	}
}

func TestHandler_UpdateSupply_ExplicitIsPartialPrice_DoesNotFetchCurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		getSupplyFn: func(_ context.Context, _ int64) (*domain.Supply, error) {
			t.Fatalf("GetSupply should not be called when is_partial_price is explicit")
			return nil, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, _ := newHandlerJSONContext(http.MethodPut, "/api/v1/supplies/42", `{
		"project_id": 10,
		"name": "Urea",
		"price": "125.00",
		"is_partial_price": false,
		"unit_id": 2,
		"category_id": 5,
		"type_id": 3
	}`)
	ctx.Params = gin.Params{{Key: "supply_id", Value: "42"}}

	h.UpdateSupply(ctx)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
	assert.Empty(t, stub.getSupplyCalls)
	if assert.Len(t, stub.updateSupplyCalls, 1) {
		assert.False(t, stub.updateSupplyCalls[0].IsPartialPrice)
	}
}

func TestHandler_UpdateSuppliesBulk_OmittedIsPartialPrice_UsesSingleBatchLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		getSuppliesByIDsFn: func(_ context.Context, _ []int64) (map[int64]domain.Supply, error) {
			return map[int64]domain.Supply{
				1: {ID: 1, IsPartialPrice: true},
				2: {ID: 2, IsPartialPrice: false},
			}, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, _ := newHandlerJSONContext(http.MethodPut, "/api/v1/supplies/bulk", `[
		{
			"id": 1,
			"project_id": 10,
			"name": "A",
			"price": "10",
			"unit_id": 2,
			"category_id": 5,
			"type_id": 3
		},
		{
			"id": 2,
			"project_id": 10,
			"name": "B",
			"price": "20",
			"unit_id": 2,
			"category_id": 5,
			"type_id": 3
		},
		{
			"id": 3,
			"project_id": 10,
			"name": "C",
			"price": "30",
			"is_partial_price": false,
			"unit_id": 2,
			"category_id": 5,
			"type_id": 3
		}
	]`)

	h.UpdateSuppliesBulk(ctx)

	assert.Equal(t, http.StatusNoContent, ctx.Writer.Status())
	if assert.Len(t, stub.getSuppliesIDsCalls, 1) {
		assert.ElementsMatch(t, []int64{1, 2}, stub.getSuppliesIDsCalls[0])
	}
	assert.Empty(t, stub.getSupplyCalls)
	if assert.Len(t, stub.updateBulkCalls, 1) {
		gotByID := make(map[int64]domain.Supply, len(stub.updateBulkCalls[0]))
		for i := range stub.updateBulkCalls[0] {
			gotByID[stub.updateBulkCalls[0][i].ID] = stub.updateBulkCalls[0][i]
		}
		assert.True(t, gotByID[1].IsPartialPrice)
		assert.False(t, gotByID[2].IsPartialPrice)
		assert.False(t, gotByID[3].IsPartialPrice)
	}
}

func TestHandler_UpdateSuppliesBulk_OmittedIsPartialPrice_MissingSupplyReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		getSuppliesByIDsFn: func(_ context.Context, _ []int64) (map[int64]domain.Supply, error) {
			return map[int64]domain.Supply{}, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, _ := newHandlerJSONContext(http.MethodPut, "/api/v1/supplies/bulk", `[
		{
			"id": 99,
			"project_id": 10,
			"name": "A",
			"price": "10",
			"unit_id": 2,
			"category_id": 5,
			"type_id": 3
		}
	]`)

	h.UpdateSuppliesBulk(ctx)

	assert.Equal(t, http.StatusNotFound, ctx.Writer.Status())
	assert.Len(t, stub.getSuppliesIDsCalls, 1)
	assert.Empty(t, stub.updateBulkCalls)
}
