package lot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	domain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
)

func TestLotIDActionHandlersCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name     string
		run      func(*Handler, *gin.Context)
		wantCall string
	}{
		{
			name: "legacy delete calls legacy archive alias",
			run: func(h *Handler, c *gin.Context) {
				h.DeleteLot(c)
			},
			wantCall: "delete:42",
		},
		{
			name: "archive calls archive usecase",
			run: func(h *Handler, c *gin.Context) {
				h.ArchiveLot(c)
			},
			wantCall: "archive:42",
		},
		{
			name: "restore calls restore usecase",
			run: func(h *Handler, c *gin.Context) {
				h.RestoreLot(c)
			},
			wantCall: "restore:42",
		},
		{
			name: "hard delete calls hard-delete usecase",
			run: func(h *Handler, c *gin.Context) {
				h.HardDeleteLot(c)
			},
			wantCall: "hard:42",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucs := &lotActionUseCasesSpy{}
			h := &Handler{ucs: ucs}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/lots/42/action", nil)
			c.Params = gin.Params{{Key: "lot_id", Value: "42"}}

			tt.run(h, c)

			if c.Writer.Status() != http.StatusNoContent {
				t.Fatalf("expected status 204, got %d body=%s", c.Writer.Status(), rec.Body.String())
			}
			if ucs.call != tt.wantCall {
				t.Fatalf("expected call %q, got %q", tt.wantCall, ucs.call)
			}
		})
	}
}

type lotActionUseCasesSpy struct {
	call string
}

func (s *lotActionUseCasesSpy) CreateLot(context.Context, *domain.Lot) (int64, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) GetLot(context.Context, int64) (*domain.Lot, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) UpdateLot(context.Context, *domain.Lot) error {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) UpdateLotTons(context.Context, int64, decimal.Decimal) error {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) ArchiveLot(_ context.Context, id int64) error {
	s.call = "archive:42"
	if id != 42 {
		s.call = "archive:unexpected"
	}
	return nil
}

func (s *lotActionUseCasesSpy) RestoreLot(_ context.Context, id int64) error {
	s.call = "restore:42"
	if id != 42 {
		s.call = "restore:unexpected"
	}
	return nil
}

func (s *lotActionUseCasesSpy) HardDeleteLot(_ context.Context, id int64) error {
	s.call = "hard:42"
	if id != 42 {
		s.call = "hard:unexpected"
	}
	return nil
}

func (s *lotActionUseCasesSpy) ListArchivedLots(context.Context, int, int) ([]domain.Lot, int64, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) DeleteLot(_ context.Context, id int64) error {
	s.call = "delete:42"
	if id != 42 {
		s.call = "delete:unexpected"
	}
	return nil
}

func (s *lotActionUseCasesSpy) ListLotsByField(context.Context, int64) ([]domain.Lot, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) ListLotsByProject(context.Context, int64) ([]domain.Lot, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) GetMetrics(context.Context, domain.LotListFilter) (*domain.LotMetrics, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) ListLots(context.Context, domain.LotListFilter, int, int) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error) {
	panic("not implemented")
}

func (s *lotActionUseCasesSpy) ExportLots(context.Context, domain.LotListFilter, int, int) ([]byte, error) {
	panic("not implemented")
}
