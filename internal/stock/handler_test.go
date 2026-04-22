package stock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctxkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type stockHandlerUseCasesStub struct {
	getStocksSummaryFn func(context.Context, int64, time.Time) ([]*domain.Stock, error)
	createStockCountFn func(context.Context, int64, int64, *domain.StockCount) (int64, error)

	getStocksSummaryCalls []struct {
		projectID  int64
		cutoffDate time.Time
	}
	createStockCountCalls []struct {
		projectID int64
		supplyID  int64
		count     *domain.StockCount
	}
}

func (s *stockHandlerUseCasesStub) GetStocksSummary(
	ctx context.Context,
	projectID int64,
	cutoffDate time.Time,
) ([]*domain.Stock, error) {
	s.getStocksSummaryCalls = append(s.getStocksSummaryCalls, struct {
		projectID  int64
		cutoffDate time.Time
	}{
		projectID:  projectID,
		cutoffDate: cutoffDate,
	})
	if s.getStocksSummaryFn != nil {
		return s.getStocksSummaryFn(ctx, projectID, cutoffDate)
	}
	return nil, nil
}

func (s *stockHandlerUseCasesStub) GetStockBySupplyID(context.Context, int64, int64, time.Time) (*domain.Stock, error) {
	return nil, nil
}

func (s *stockHandlerUseCasesStub) CreateStockCount(
	ctx context.Context,
	projectID int64,
	supplyID int64,
	count *domain.StockCount,
) (int64, error) {
	s.createStockCountCalls = append(s.createStockCountCalls, struct {
		projectID int64
		supplyID  int64
		count     *domain.StockCount
	}{
		projectID: projectID,
		supplyID:  supplyID,
		count:     count,
	})
	if s.createStockCountFn != nil {
		return s.createStockCountFn(ctx, projectID, supplyID, count)
	}
	return 0, nil
}

func (s *stockHandlerUseCasesStub) ExportStocksByProject(context.Context, int64) ([]byte, error) {
	return nil, nil
}

func TestHandler_GetStocksSummary_UsesCutoffDateAndSupplySummaryShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	countedAt := time.Date(2026, 4, 20, 18, 0, 0, 0, time.UTC)
	stub := &stockHandlerUseCasesStub{
		getStocksSummaryFn: func(context.Context, int64, time.Time) ([]*domain.Stock, error) {
			return []*domain.Stock{
				{
					ID:                9,
					ProjectID:         7,
					EntryStock:        decimal.NewFromInt(100),
					OutStock:          decimal.NewFromInt(12),
					Consumed:          decimal.NewFromInt(5),
					StockUnits:        decimal.NewFromInt(83),
					RealStockUnits:    decimal.NewFromInt(80),
					HasRealStockCount: true,
					LastCountAt:       &countedAt,
					Supply: &supplydomain.Supply{
						ID:           9,
						Name:         "Urea",
						UnitID:       1,
						UnitName:     "kg",
						Price:        decimal.NewFromFloat(3.5),
						CategoryName: "Fertilizantes",
					},
				},
			}, nil
		},
	}
	h := &Handler{ucs: stub}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/projects/7/stocks/summary?cutoff_date=2026-04-21", nil)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "project_id", Value: "7"}}

	h.getStocksSummary(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if len(stub.getStocksSummaryCalls) != 1 {
		t.Fatalf("expected one use case call, got %d", len(stub.getStocksSummaryCalls))
	}
	call := stub.getStocksSummaryCalls[0]
	if call.projectID != 7 {
		t.Fatalf("expected project_id 7, got %d", call.projectID)
	}
	if got := call.cutoffDate.Format("2006-01-02"); got != "2026-04-21" {
		t.Fatalf("expected cutoff date 2026-04-21, got %s", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"supply_id":9`) {
		t.Fatalf("expected response to include supply_id, got %s", body)
	}
	if !strings.Contains(body, `"out_stock":"12.00"`) {
		t.Fatalf("expected response to include out_stock, got %s", body)
	}
	if strings.Contains(body, "investor_name") || strings.Contains(body, "close_date") {
		t.Fatalf("response should not expose legacy fields, got %s", body)
	}
}

func TestHandler_CreateStockCount_UsesActorFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &stockHandlerUseCasesStub{
		createStockCountFn: func(context.Context, int64, int64, *domain.StockCount) (int64, error) {
			return 55, nil
		},
	}
	h := &Handler{ucs: stub}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(
		http.MethodPost,
		"/projects/7/supplies/9/stock-counts",
		strings.NewReader(`{"counted_units":"82","counted_at":"2026-04-22T15:04:05Z","note":"Conteo general"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.Actor, "auditor@ponti.test"))
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "project_id", Value: "7"},
		{Key: "supply_id", Value: "9"},
	}

	h.CreateStockCount(ctx)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if len(stub.createStockCountCalls) != 1 {
		t.Fatalf("expected one create call, got %d", len(stub.createStockCountCalls))
	}
	call := stub.createStockCountCalls[0]
	if call.projectID != 7 || call.supplyID != 9 {
		t.Fatalf("unexpected ids: %+v", call)
	}
	if call.count == nil {
		t.Fatal("expected count payload")
	}
	if call.count.CreatedBy == nil || *call.count.CreatedBy != "auditor@ponti.test" {
		t.Fatalf("expected actor in CreatedBy, got %+v", call.count.CreatedBy)
	}
	if call.count.Note != "Conteo general" {
		t.Fatalf("expected note to be forwarded, got %q", call.count.Note)
	}
	if !call.count.CountedUnits.Equal(decimal.NewFromInt(82)) {
		t.Fatalf("expected counted units 82, got %s", call.count.CountedUnits)
	}

	var resp struct {
		ID      int64  `json:"id"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.ID != 55 {
		t.Fatalf("expected response id 55, got %d", resp.ID)
	}
}
