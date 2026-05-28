package dataintegrity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
)

type dataIntegrityHandlerUseCasesStub struct {
	filters                []domain.CostsCheckFilter
	tentativePricesFilters []domain.TentativePricesFilter
}

func (s *dataIntegrityHandlerUseCasesStub) CheckCostsCoherence(_ context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error) {
	s.filters = append(s.filters, filter)
	return &domain.IntegrityReport{
		Checks: []domain.IntegrityCheck{
			{
				ControlNumber:      1,
				DataToVerify:       "costos",
				Description:        "descripcion",
				ControlRule:        "rule",
				SystemCalculation:  "system",
				SystemValue:        decimal.NewFromInt(10),
				SystemSource:       "dashboard",
				SystemMeaning:      "meaning",
				RecalcACalculation: "recalc",
				RecalcAValue:       decimal.NewFromInt(10),
				RecalcASource:      "report",
				RecalcAMeaning:     "meaning",
				DifferenceA:        decimal.Zero,
				Status:             "OK",
				Tolerance:          decimal.Zero,
			},
		},
	}, nil
}

func (s *dataIntegrityHandlerUseCasesStub) GetTentativePrices(_ context.Context, filter domain.TentativePricesFilter) (*domain.TentativePricesReport, error) {
	s.tentativePricesFilters = append(s.tentativePricesFilters, filter)
	return &domain.TentativePricesReport{
		Count: 1,
		Items: []domain.TentativePriceItem{
			{
				SupplyID:     123,
				Name:         "Glifosato",
				CategoryName: "Herbicidas",
				Price:        decimal.RequireFromString("4.5"),
			},
		},
	}, nil
}

func newDataIntegrityHandlerContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(""))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_CheckCostsCoherence_ParsesProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &dataIntegrityHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newDataIntegrityHandlerContext(http.MethodGet, "/api/v1/data-integrity/costs-check?project_id=42")

	h.CheckCostsCoherence(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.filters) != 1 || stub.filters[0].ProjectID == nil || *stub.filters[0].ProjectID != 42 {
		t.Fatalf("expected project_id 42, got %#v", stub.filters)
	}
}

func TestHandler_CheckCostsCoherence_RequiresProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &Handler{ucs: &dataIntegrityHandlerUseCasesStub{}}
	ctx, _ := newDataIntegrityHandlerContext(http.MethodGet, "/api/v1/data-integrity/costs-check")

	h.CheckCostsCoherence(ctx)

	if ctx.Writer.Status() != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.Writer.Status())
	}
}

func TestHandler_GetTentativePrices_ParsesWorkspaceFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &dataIntegrityHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, rec := newDataIntegrityHandlerContext(http.MethodGet, "/api/v1/data-integrity/tentative-prices?customer_id=1&project_id=2&campaign_id=3&field_id=4")

	h.GetTentativePrices(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, ctx.Writer.Status(), rec.Body.String())
	}
	if len(stub.tentativePricesFilters) != 1 {
		t.Fatalf("expected one tentative prices filter, got %#v", stub.tentativePricesFilters)
	}
	filter := stub.tentativePricesFilters[0]
	if filter.CustomerID == nil || *filter.CustomerID != 1 ||
		filter.ProjectID == nil || *filter.ProjectID != 2 ||
		filter.CampaignID == nil || *filter.CampaignID != 3 ||
		filter.FieldID == nil || *filter.FieldID != 4 {
		t.Fatalf("unexpected filter: %#v", filter)
	}
	if !strings.Contains(rec.Body.String(), `"price":"4.50"`) {
		t.Fatalf("expected formatted price in response, got %s", rec.Body.String())
	}
}
