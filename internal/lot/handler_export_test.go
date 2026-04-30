package lot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
)

type exportLotsUseCasesStub struct {
	filter   domain.LotListFilter
	page     int
	pageSize int
}

func (s *exportLotsUseCasesStub) CreateLot(context.Context, *domain.Lot) (int64, error) {
	return 0, nil
}

func (s *exportLotsUseCasesStub) GetLot(context.Context, int64) (*domain.Lot, error) {
	return nil, nil
}

func (s *exportLotsUseCasesStub) UpdateLot(context.Context, *domain.Lot) error {
	return nil
}

func (s *exportLotsUseCasesStub) UpdateLotTons(context.Context, int64, decimal.Decimal) error {
	return nil
}

func (s *exportLotsUseCasesStub) DeleteLot(context.Context, int64) error {
	return nil
}

func (s *exportLotsUseCasesStub) ListLotsByField(context.Context, int64) ([]domain.Lot, error) {
	return nil, nil
}

func (s *exportLotsUseCasesStub) ListLotsByProject(context.Context, int64) ([]domain.Lot, error) {
	return nil, nil
}

func (s *exportLotsUseCasesStub) ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error) {
	return nil, nil
}

func (s *exportLotsUseCasesStub) ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error) {
	return nil, nil
}

func (s *exportLotsUseCasesStub) GetMetrics(context.Context, int64, int64, int64) (*domain.LotMetrics, error) {
	return nil, nil
}

func (s *exportLotsUseCasesStub) ListLots(context.Context, domain.LotListFilter, int, int) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error) {
	return nil, 0, decimal.Zero, decimal.Zero, nil
}

func (s *exportLotsUseCasesStub) ExportLots(_ context.Context, filter domain.LotListFilter, page, pageSize int) ([]byte, error) {
	s.filter = filter
	s.page = page
	s.pageSize = pageSize
	return []byte("xlsx"), nil
}

func TestExportLotsIgnoresRequestedPageAndUsesBackendLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &exportLotsUseCasesStub{}
	handler := &Handler{ucs: stub}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/v1/lots/export?project_id=30&page=7&per_page=10",
		nil,
	)

	handler.ExportLots(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if stub.page != 1 {
		t.Fatalf("page = %d, want 1", stub.page)
	}
	if stub.pageSize != maxLotExportPageSize {
		t.Fatalf("pageSize = %d, want %d", stub.pageSize, maxLotExportPageSize)
	}
	if stub.filter.ProjectID == nil || *stub.filter.ProjectID != 30 {
		t.Fatalf("project filter = %#v, want 30", stub.filter.ProjectID)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("content type = %q", got)
	}
}
