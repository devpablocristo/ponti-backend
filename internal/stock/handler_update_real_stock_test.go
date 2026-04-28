package stock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/devpablocristo/core/security/go/contextkeys"
	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type updateRealStockFakeUseCases struct {
	UseCasesPort
	stock        *domain.Stock
	updatedStock *domain.Stock
	updateCalled bool
}

func (f *updateRealStockFakeUseCases) GetStockByID(context.Context, int64) (*domain.Stock, error) {
	return f.stock, nil
}

func (f *updateRealStockFakeUseCases) UpdateRealStockUnits(_ context.Context, _ int64, stock *domain.Stock) error {
	f.updateCalled = true
	f.updatedStock = stock
	return nil
}

func TestHandler_UpdateRealStockRejectsStockFromOtherProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fakeUC := &updateRealStockFakeUseCases{
		stock: &domain.Stock{
			ID:      10,
			Project: &projectdomain.Project{ID: 2},
		},
	}
	handler := &Handler{ucs: fakeUC}
	ctx, recorder := newUpdateRealStockTestContext(
		t,
		"1",
		"10",
		`{"real_stock_units":"7"}`,
	)

	handler.UpdateRealStock(ctx)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.False(t, fakeUC.updateCalled)
}

func TestHandler_UpdateRealStockUsesClientUpdatedAtForOptimisticLock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serverVersion := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	clientVersion := serverVersion.Add(-time.Minute)
	fakeUC := &updateRealStockFakeUseCases{
		stock: &domain.Stock{
			ID:      10,
			Project: &projectdomain.Project{ID: 1},
			Base: shareddomain.Base{
				UpdatedAt: serverVersion,
			},
		},
	}
	handler := &Handler{ucs: fakeUC}
	ctx, recorder := newUpdateRealStockTestContext(
		t,
		"1",
		"10",
		`{"real_stock_units":"7","updated_at":"`+clientVersion.Format(time.RFC3339)+`"}`,
	)

	handler.UpdateRealStock(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.True(t, fakeUC.updateCalled)
	require.NotNil(t, fakeUC.updatedStock)
	assert.True(t, fakeUC.updatedStock.UpdatedAt.Equal(clientVersion))
	assert.True(t, fakeUC.updatedStock.RealStockUnits.Equal(decimal.NewFromInt(7)))
	assert.True(t, fakeUC.updatedStock.HasRealStockCount)
	require.NotNil(t, fakeUC.updatedStock.UpdatedBy)
	assert.Equal(t, "user@example.com", *fakeUC.updatedStock.UpdatedBy)
}

func newUpdateRealStockTestContext(t *testing.T, projectID, stockID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPut, "/projects/"+projectID+"/stocks/real-stock/"+stockID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.Actor, "user@example.com"))
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "project_id", Value: projectID},
		{Key: "stock_id", Value: stockID},
	}
	return ctx, recorder
}
