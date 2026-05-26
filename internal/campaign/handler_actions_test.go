package campaign

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
)

func TestCampaignIDActionHandlersCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name     string
		run      func(*Handler, *gin.Context)
		wantCall string
	}{
		{
			name: "archive calls archive usecase",
			run: func(h *Handler, c *gin.Context) {
				h.ArchiveCampaign(c)
			},
			wantCall: "archive:42",
		},
		{
			name: "restore calls restore usecase",
			run: func(h *Handler, c *gin.Context) {
				h.RestoreCampaign(c)
			},
			wantCall: "restore:42",
		},
		{
			name: "hard delete calls hard-delete usecase",
			run: func(h *Handler, c *gin.Context) {
				h.HardDeleteCampaign(c)
			},
			wantCall: "hard:42",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucs := &campaignActionUseCasesSpy{}
			h := &Handler{ucs: ucs}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/campaigns/42/action", nil)
			c.Params = gin.Params{{Key: "campaign_id", Value: "42"}}

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

type campaignActionUseCasesSpy struct {
	call string
}

func (s *campaignActionUseCasesSpy) CreateCampaign(context.Context, *domain.Campaign) (int64, error) {
	panic("not implemented")
}

func (s *campaignActionUseCasesSpy) ListCampaigns(context.Context, int64, string) ([]domain.Campaign, error) {
	panic("not implemented")
}

func (s *campaignActionUseCasesSpy) ListArchivedCampaigns(context.Context, int, int) ([]domain.Campaign, int64, error) {
	panic("not implemented")
}

func (s *campaignActionUseCasesSpy) GetCampaign(context.Context, int64) (*domain.Campaign, error) {
	panic("not implemented")
}

func (s *campaignActionUseCasesSpy) UpdateCampaign(context.Context, *domain.Campaign) error {
	panic("not implemented")
}

func (s *campaignActionUseCasesSpy) ArchiveCampaign(_ context.Context, id int64) error {
	s.call = "archive:42"
	if id != 42 {
		s.call = "archive:unexpected"
	}
	return nil
}

func (s *campaignActionUseCasesSpy) RestoreCampaign(_ context.Context, id int64) error {
	s.call = "restore:42"
	if id != 42 {
		s.call = "restore:unexpected"
	}
	return nil
}

func (s *campaignActionUseCasesSpy) HardDeleteCampaign(_ context.Context, id int64) error {
	s.call = "hard:42"
	if id != 42 {
		s.call = "hard:unexpected"
	}
	return nil
}
