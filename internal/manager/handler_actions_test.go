package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
)

func TestManagerIDActionHandlersCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name     string
		run      func(*Handler, *gin.Context)
		wantCall string
	}{
		{
			name: "hard delete calls hard-delete usecase",
			run: func(h *Handler, c *gin.Context) {
				h.HardDeleteManager(c)
			},
			wantCall: "hard:42",
		},
		{
			name: "archive calls archive usecase",
			run: func(h *Handler, c *gin.Context) {
				h.ArchiveManager(c)
			},
			wantCall: "archive:42",
		},
		{
			name: "restore calls restore usecase",
			run: func(h *Handler, c *gin.Context) {
				h.RestoreManager(c)
			},
			wantCall: "restore:42",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucs := &managerActionUseCasesSpy{}
			h := &Handler{ucs: ucs}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/managers/42/action", nil)
			c.Params = gin.Params{{Key: "manager_id", Value: "42"}}

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

type managerActionUseCasesSpy struct {
	call string
}

func (s *managerActionUseCasesSpy) CreateManager(context.Context, *domain.Manager) (int64, error) {
	panic("not implemented")
}

func (s *managerActionUseCasesSpy) ListManagers(context.Context, int, int) ([]domain.Manager, int64, error) {
	panic("not implemented")
}

func (s *managerActionUseCasesSpy) ListArchivedManagers(context.Context, int, int) ([]domain.Manager, int64, error) {
	panic("not implemented")
}

func (s *managerActionUseCasesSpy) GetManager(context.Context, int64) (*domain.Manager, error) {
	panic("not implemented")
}

func (s *managerActionUseCasesSpy) UpdateManager(context.Context, *domain.Manager) error {
	panic("not implemented")
}

func (s *managerActionUseCasesSpy) ArchiveManager(_ context.Context, id int64) error {
	s.call = "archive:42"
	if id != 42 {
		s.call = "archive:unexpected"
	}
	return nil
}

func (s *managerActionUseCasesSpy) RestoreManager(_ context.Context, id int64) error {
	s.call = "restore:42"
	if id != 42 {
		s.call = "restore:unexpected"
	}
	return nil
}

func (s *managerActionUseCasesSpy) HardDeleteManager(_ context.Context, id int64) error {
	s.call = "hard:42"
	if id != 42 {
		s.call = "hard:unexpected"
	}
	return nil
}
