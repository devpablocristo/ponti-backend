package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
)

type dashboardHandlerUseCasesStub struct {
	filters []domain.DashboardFilter
}

func (s *dashboardHandlerUseCasesStub) GetDashboard(_ context.Context, filter domain.DashboardFilter) (*domain.DashboardData, error) {
	s.filters = append(s.filters, filter)
	return &domain.DashboardData{}, nil
}

func newDashboardHandlerContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(""))
	ctx.Request = req
	return ctx, rec
}

func TestHandler_GetDashboard_ParsesWorkspaceFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &dashboardHandlerUseCasesStub{}
	h := &Handler{ucs: stub}
	ctx, _ := newDashboardHandlerContext(http.MethodGet, "/api/v1/dashboard?customer_id=1&project_id=2&campaign_id=3&field_id=4")

	h.GetDashboard(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.filters) != 1 {
		t.Fatalf("expected one dashboard call, got %#v", stub.filters)
	}
	got := stub.filters[0]
	if got.CustomerID == nil || *got.CustomerID != 1 || got.ProjectID == nil || *got.ProjectID != 2 || got.CampaignID == nil || *got.CampaignID != 3 || got.FieldID == nil || *got.FieldID != 4 {
		t.Fatalf("unexpected dashboard filter: %#v", got)
	}
}
