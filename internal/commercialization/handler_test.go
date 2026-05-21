package commercialization

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	domain "github.com/devpablocristo/ponti-backend/internal/commercialization/usecases/domain"
)

type commercializationHandlerUseCasesStub struct {
	listProjectCalls []int64
	bulkCalls        [][]domain.CropCommercialization
}

func (s *commercializationHandlerUseCasesStub) CreateOrUpdateBulk(_ context.Context, items []domain.CropCommercialization) error {
	s.bulkCalls = append(s.bulkCalls, items)
	return nil
}

func (s *commercializationHandlerUseCasesStub) ListByProject(_ context.Context, projectID int64) ([]domain.CropCommercialization, error) {
	s.listProjectCalls = append(s.listProjectCalls, projectID)
	return []domain.CropCommercialization{{ProjectID: projectID, CropID: 1, BoardPrice: decimal.NewFromInt(100)}}, nil
}

func newCommercializationHandlerContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.Actor, "tester@example.com"))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_CommercializationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &commercializationHandlerUseCasesStub{}
	h := &Handler{ucs: stub}

	listCtx, _ := newCommercializationHandlerContext(http.MethodGet, "/api/v1/projects/10/commercializations", "")
	listCtx.Params = gin.Params{{Key: "project_id", Value: "10"}}
	h.ListByProject(listCtx)
	if listCtx.Writer.Status() != http.StatusOK || len(stub.listProjectCalls) != 1 || stub.listProjectCalls[0] != 10 {
		t.Fatalf("expected list project 10, status %d calls %#v", listCtx.Writer.Status(), stub.listProjectCalls)
	}

	body := `{"values":[{"id":7,"crop_id":3,"board_price":"100","freight_cost":"10","commercial_cost":"5"}]}`
	bulkCtx, _ := newCommercializationHandlerContext(http.MethodPost, "/api/v1/projects/10/commercializations", body)
	bulkCtx.Params = gin.Params{{Key: "project_id", Value: "10"}}
	h.CreateOrUpdateBulk(bulkCtx)
	if bulkCtx.Writer.Status() != http.StatusNoContent || len(stub.bulkCalls) != 1 || len(stub.bulkCalls[0]) != 1 {
		t.Fatalf("expected one bulk item, status %d calls %#v", bulkCtx.Writer.Status(), stub.bulkCalls)
	}
	got := stub.bulkCalls[0][0]
	if got.ProjectID != 10 || got.ID != 7 || got.CropID != 3 || got.CreatedBy == nil || *got.CreatedBy != "tester@example.com" {
		t.Fatalf("unexpected bulk item: %#v", got)
	}
}
