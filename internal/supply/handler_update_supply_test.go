package supply

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/devpablocristo/core/errors/go/domainerr"

	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	domain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type handlerUseCasesStub struct {
	getSupplyFn             func(ctx context.Context, id int64) (*domain.Supply, error)
	getSuppliesByIDsFn      func(ctx context.Context, ids []int64) (map[int64]domain.Supply, error)
	updateSupplyFn          func(ctx context.Context, s *domain.Supply) error
	updateSuppliesFn        func(ctx context.Context, supplies []domain.Supply) error
	createSupplyMovementFn  func(ctx context.Context, movement *domain.SupplyMovement) (int64, error)
	validateMovementFn      func(ctx context.Context, movement *domain.SupplyMovement) error
	importSupplyMovementsFn func(ctx context.Context, movements []*domain.SupplyMovement) ([]int64, []SupplyMovementImportFailure, error)
	getSupplyCalls          []int64
	getSuppliesIDsCalls     [][]int64
	updateSupplyCalls       []domain.Supply
	updateBulkCalls         [][]domain.Supply
	importCalls             [][]*domain.SupplyMovement
}

func (s *handlerUseCasesStub) CreateSupply(context.Context, *domain.Supply) (int64, error) {
	return 0, nil
}

func (s *handlerUseCasesStub) CreatePendingSupply(context.Context, int64, string) (*domain.Supply, bool, error) {
	return nil, false, nil
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

func (s *handlerUseCasesStub) CompletePendingSupply(ctx context.Context, supply *domain.Supply) error {
	return s.UpdateSupply(ctx, supply)
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

func (s *handlerUseCasesStub) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if s.createSupplyMovementFn != nil {
		return s.createSupplyMovementFn(ctx, movement)
	}
	return 0, nil
}

func (s *handlerUseCasesStub) ValidateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) error {
	if s.validateMovementFn != nil {
		return s.validateMovementFn(ctx, movement)
	}
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

func (s *handlerUseCasesStub) ArchiveSupply(ctx context.Context, id int64) error {
	return nil
}

func (s *handlerUseCasesStub) RestoreSupply(ctx context.Context, id int64) error {
	return nil
}

func newHandlerJSONContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxkeys.Actor, "test@example.com"))
	return ctx, rec
}

func (s *handlerUseCasesStub) ImportSupplyMovements(ctx context.Context, movements []*domain.SupplyMovement) ([]int64, []SupplyMovementImportFailure, error) {
	callCopy := append([]*domain.SupplyMovement(nil), movements...)
	s.importCalls = append(s.importCalls, callCopy)
	if s.importSupplyMovementsFn != nil {
		return s.importSupplyMovementsFn(ctx, movements)
	}
	return nil, nil, nil
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

func TestHandler_ImportSupplyMovements_ReturnsRowIndexFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		importSupplyMovementsFn: func(_ context.Context, _ []*domain.SupplyMovement) ([]int64, []SupplyMovementImportFailure, error) {
			return nil, []SupplyMovementImportFailure{{
				Index:    1,
				RowIndex: 3,
				SupplyID: 99,
				Code:     "duplicate_request",
				Message:  "El remito R-1 ya contiene el insumo 99 dentro del request",
			}}, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements/import", `{
		"mode": "strict",
		"items": [
			{
				"quantity": "10",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "R-1",
				"project_destination_id": 0,
				"supply_id": 10,
				"investor_id": 11,
				"provider": { "id": 5, "name": "Provider 5" }
			},
			{
				"quantity": "10",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "R-1",
				"project_destination_id": 0,
				"supply_id": 99,
				"investor_id": 11,
				"provider": { "id": 5, "name": "Provider 5" }
			}
		]
	}`)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}

	h.ImportSupplyMovements(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Success  bool `json:"success"`
		Failures []struct {
			Index    int    `json:"index"`
			RowIndex int    `json:"row_index"`
			SupplyID int64  `json:"supply_id"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"failures"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	if assert.Len(t, resp.Failures, 1) {
		assert.Equal(t, 1, resp.Failures[0].Index)
		assert.Equal(t, 3, resp.Failures[0].RowIndex)
		assert.Equal(t, int64(99), resp.Failures[0].SupplyID)
	}
}

func TestHandler_ImportSupplyMovements_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		importSupplyMovementsFn: func(_ context.Context, movements []*domain.SupplyMovement) ([]int64, []SupplyMovementImportFailure, error) {
			assert.Len(t, movements, 2)
			assert.Equal(t, int64(18), movements[0].ProjectId)
			return []int64{101, 102}, nil, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements/import", `{
		"mode": "strict",
		"items": [
			{
				"quantity": "10",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "R-1",
				"project_destination_id": 0,
				"supply_id": 10,
				"investor_id": 11,
				"provider": { "id": 5, "name": "Provider 5" }
			},
			{
				"quantity": "11",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-05T00:00:00Z",
				"reference_number": "R-2",
				"project_destination_id": 0,
				"supply_id": 11,
				"investor_id": 12,
				"provider": { "id": 6, "name": "Provider 6" }
			}
		]
	}`)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}

	h.ImportSupplyMovements(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Success bool `json:"success"`
		Applied int  `json:"applied"`
		Failed  int  `json:"failed"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, 2, resp.Applied)
	assert.Equal(t, 0, resp.Failed)
}

func TestHandler_ImportSupplyMovements_InitialValidationUsesRowIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{ucs: &handlerUseCasesStub{}}
	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements/import", `{
		"mode": "strict",
		"items": [
			{
				"quantity": "10",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "R-1",
				"project_destination_id": 0,
				"supply_id": 10,
				"investor_id": 11,
				"provider": { "id": 5, "name": "Provider 5" }
			},
			{
				"quantity": "10",
				"movement_type": "Invalido",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "R-2",
				"project_destination_id": 0,
				"supply_id": 11,
				"investor_id": 12,
				"provider": { "id": 6, "name": "Provider 6" }
			}
		]
	}`)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}

	h.ImportSupplyMovements(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Failures []struct {
			Index    int `json:"index"`
			RowIndex int `json:"row_index"`
		} `json:"failures"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	if assert.Len(t, resp.Failures, 1) {
		assert.Equal(t, 1, resp.Failures[0].Index)
		assert.Equal(t, 3, resp.Failures[0].RowIndex)
	}
}

func TestHandler_ImportSupplyMovements_InvalidUserIDReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{ucs: &handlerUseCasesStub{}}
	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements/import", `{"mode":"strict","items":[]}`)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), ctxkeys.Actor, ""))

	h.ImportSupplyMovements(ctx)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "FORBIDDEN", resp.Code)
}

func TestHandler_CreateSupplyMovement_StrictReturnsDuplicateFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		validateMovementFn: func(_ context.Context, movement *domain.SupplyMovement) error {
			if movement.ReferenceNumber == "REM-EXCEL" && movement.Supply != nil && movement.Supply.ID == 10 {
				return domainerr.Conflict("El remito REM-EXCEL ya tiene el insumo 10 cargado")
			}
			return nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements", `{
		"mode": "strict",
		"items": [
			{
				"quantity": "2",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "REM-EXCEL",
				"project_destination_id": 0,
				"supply_id": 10,
				"investor_id": 11,
				"provider": { "id": 5, "name": "Provider 5" }
			}
		]
	}`)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}

	h.CreateSupplyMovement(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Success  bool `json:"success"`
		Applied  int  `json:"applied"`
		Failed   int  `json:"failed"`
		Failures []struct {
			Index    int    `json:"index"`
			RowIndex int    `json:"row_index"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"failures"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, 0, resp.Applied)
	assert.Equal(t, 1, resp.Failed)
	if assert.Len(t, resp.Failures, 1) {
		assert.Equal(t, 0, resp.Failures[0].Index)
		assert.Equal(t, 2, resp.Failures[0].RowIndex)
		assert.Equal(t, "validation_error", resp.Failures[0].Code)
		assert.Equal(t, "El remito REM-EXCEL ya tiene el insumo 10 cargado", resp.Failures[0].Message)
	}
}

func TestHandler_ImportSupplyMovements_ExceedsMaxItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{}
	h := &Handler{ucs: stub}

	items := make([]string, 501)
	for i := range items {
		items[i] = `{
			"quantity": "1",
			"movement_type": "Remito oficial",
			"movement_date": "2026-03-04T00:00:00Z",
			"reference_number": "REM-1",
			"supply_id": 10,
			"investor_id": 5,
			"provider": {"id": 1, "name": "P"}
		}`
	}
	body := `{"items": [` + strings.Join(items, ",") + `]}`

	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements/import", body)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}

	h.ImportSupplyMovements(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Message, "500")
	assert.Contains(t, resp.Message, "501")
}

func TestHandler_ImportSupplyMovements_FailuresReturnWarningWithAccents(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &handlerUseCasesStub{
		importSupplyMovementsFn: func(_ context.Context, _ []*domain.SupplyMovement) ([]int64, []SupplyMovementImportFailure, error) {
			return nil, []SupplyMovementImportFailure{
				{Index: 0, RowIndex: 2, SupplyID: 10, Code: "duplicate_db", Message: "duplicado"},
			}, nil
		},
	}

	h := &Handler{ucs: stub}
	ctx, rec := newHandlerJSONContext(http.MethodPost, "/api/v1/projects/18/supply-movements/import", `{
		"items": [
			{
				"quantity": "5",
				"movement_type": "Remito oficial",
				"movement_date": "2026-03-04T00:00:00Z",
				"reference_number": "REM-1",
				"supply_id": 10,
				"investor_id": 5,
				"provider": {"id": 1, "name": "P"}
			}
		]
	}`)
	ctx.Params = gin.Params{{Key: "project_id", Value: "18"}}

	h.ImportSupplyMovements(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp struct {
		Warnings []string `json:"warnings"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	if assert.Len(t, resp.Warnings, 1) {
		assert.Equal(t, "No se guardó ningún movimiento porque la importación es atómica", resp.Warnings[0])
	}
}
