package dollar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	domain "github.com/devpablocristo/ponti-backend/internal/dollar/usecases/domain"
)

type dollarHandlerUseCasesStub struct {
	listProjectCalls []int64
	bulkCalls        [][]domain.DollarAverage
}

func (s *dollarHandlerUseCasesStub) ListByProject(_ context.Context, projectID int64) ([]domain.DollarAverage, error) {
	s.listProjectCalls = append(s.listProjectCalls, projectID)
	return []domain.DollarAverage{{ProjectID: projectID, Month: "enero", StartValue: decimal.NewFromInt(1), EndValue: decimal.NewFromInt(2), AvgValue: decimal.NewFromInt(1)}}, nil
}

func (s *dollarHandlerUseCasesStub) CreateOrUpdateBulk(_ context.Context, items []domain.DollarAverage) error {
	s.bulkCalls = append(s.bulkCalls, items)
	return nil
}

func newDollarHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func TestHandler_DollarRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &dollarHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	listCtx, _ := newDollarHandlerContext(http.MethodGet, "/api/v1/projects/10/dollar-values", "")
	listCtx.Params = gin.Params{{Key: "project_id", Value: "10"}}
	h.ListByProject(listCtx)
	if listCtx.Writer.Status() != http.StatusOK || len(stub.listProjectCalls) != 1 || stub.listProjectCalls[0] != 10 {
		t.Fatalf("expected list project 10, status %d calls %#v", listCtx.Writer.Status(), stub.listProjectCalls)
	}

	body := `{"year":2026,"values":[{"month":"enero","start_value":"100","end_value":"200","average_value":"150"}]}`
	bulkCtx, _ := newDollarHandlerContext(http.MethodPut, "/api/v1/projects/10/dollar-values", body)
	bulkCtx.Params = gin.Params{{Key: "project_id", Value: "10"}}
	h.CreateorUpdateBulk(bulkCtx)
	if bulkCtx.Writer.Status() != http.StatusNoContent || len(stub.bulkCalls) != 1 || len(stub.bulkCalls[0]) != 1 {
		t.Fatalf("expected one bulk item, status %d calls %#v", bulkCtx.Writer.Status(), stub.bulkCalls)
	}
	got := stub.bulkCalls[0][0]
	if got.ProjectID != 10 || got.Year != 2026 || got.Month != "enero" {
		t.Fatalf("unexpected bulk item: %#v", got)
	}
}
