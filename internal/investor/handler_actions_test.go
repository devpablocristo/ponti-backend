package investor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
)

func TestInvestorIDActionHandlersCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name     string
		run      func(*Handler, *gin.Context)
		wantCall string
	}{
		{
			name: "hard delete calls hard-delete usecase",
			run: func(h *Handler, c *gin.Context) {
				h.HardDeleteInvestor(c)
			},
			wantCall: "hard:42",
		},
		{
			name: "archive calls archive usecase",
			run: func(h *Handler, c *gin.Context) {
				h.ArchiveInvestor(c)
			},
			wantCall: "archive:42",
		},
		{
			name: "restore calls restore usecase",
			run: func(h *Handler, c *gin.Context) {
				h.RestoreInvestor(c)
			},
			wantCall: "restore:42",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucs := &investorActionUseCasesSpy{}
			h := &Handler{ucs: ucs}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/investors/42/action", nil)
			c.Params = gin.Params{{Key: "investor_id", Value: "42"}}

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

type investorActionUseCasesSpy struct {
	call string
}

func (s *investorActionUseCasesSpy) CreateInvestor(context.Context, *domain.Investor) (int64, error) {
	panic("not implemented")
}

func (s *investorActionUseCasesSpy) ListInvestors(context.Context, int, int) ([]domain.Investor, int64, error) {
	panic("not implemented")
}

func (s *investorActionUseCasesSpy) ListArchivedInvestors(context.Context, int, int) ([]domain.Investor, int64, error) {
	panic("not implemented")
}

func (s *investorActionUseCasesSpy) GetInvestor(context.Context, int64) (*domain.Investor, error) {
	panic("not implemented")
}

func (s *investorActionUseCasesSpy) UpdateInvestor(context.Context, *domain.Investor) error {
	panic("not implemented")
}

func (s *investorActionUseCasesSpy) ArchiveInvestor(_ context.Context, id int64) error {
	s.call = "archive:42"
	if id != 42 {
		s.call = "archive:unexpected"
	}
	return nil
}

func (s *investorActionUseCasesSpy) RestoreInvestor(_ context.Context, id int64) error {
	s.call = "restore:42"
	if id != 42 {
		s.call = "restore:unexpected"
	}
	return nil
}

func (s *investorActionUseCasesSpy) HardDeleteInvestor(_ context.Context, id int64) error {
	s.call = "hard:42"
	if id != 42 {
		s.call = "hard:unexpected"
	}
	return nil
}
