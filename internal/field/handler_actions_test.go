package field

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
)

func TestFieldIDActionHandlersCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name     string
		run      func(*Handler, *gin.Context)
		wantCall string
	}{
		{
			name: "hard delete calls hard-delete usecase",
			run: func(h *Handler, c *gin.Context) {
				h.HardDeleteField(c)
			},
			wantCall: "hard:42",
		},
		{
			name: "archive calls archive usecase",
			run: func(h *Handler, c *gin.Context) {
				h.ArchiveField(c)
			},
			wantCall: "archive:42",
		},
		{
			name: "restore calls restore usecase",
			run: func(h *Handler, c *gin.Context) {
				h.RestoreField(c)
			},
			wantCall: "restore:42",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucs := &fieldActionUseCasesSpy{}
			h := &Handler{ucs: ucs}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/fields/42/action", nil)
			c.Params = gin.Params{{Key: "field_id", Value: "42"}}

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

type fieldActionUseCasesSpy struct {
	call string
}

func (s *fieldActionUseCasesSpy) CreateField(context.Context, *domain.Field) (int64, error) {
	panic("not implemented")
}

func (s *fieldActionUseCasesSpy) ListFields(context.Context, int, int) ([]domain.Field, int64, error) {
	panic("not implemented")
}

func (s *fieldActionUseCasesSpy) ListArchivedFields(context.Context, int, int) ([]domain.Field, int64, error) {
	panic("not implemented")
}

func (s *fieldActionUseCasesSpy) GetField(context.Context, int64) (*domain.Field, error) {
	panic("not implemented")
}

func (s *fieldActionUseCasesSpy) UpdateField(context.Context, *domain.Field) error {
	panic("not implemented")
}

func (s *fieldActionUseCasesSpy) ArchiveField(_ context.Context, id int64) error {
	s.call = "archive:42"
	if id != 42 {
		s.call = "archive:unexpected"
	}
	return nil
}

func (s *fieldActionUseCasesSpy) RestoreField(_ context.Context, id int64) error {
	s.call = "restore:42"
	if id != 42 {
		s.call = "restore:unexpected"
	}
	return nil
}

func (s *fieldActionUseCasesSpy) HardDeleteField(_ context.Context, id int64) error {
	s.call = "hard:42"
	if id != 42 {
		s.call = "hard:unexpected"
	}
	return nil
}
