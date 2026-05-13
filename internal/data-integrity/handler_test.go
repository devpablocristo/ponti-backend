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
	filters []domain.CostsCheckFilter
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
