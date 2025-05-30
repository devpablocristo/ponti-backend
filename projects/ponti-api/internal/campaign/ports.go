package campaign

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type UseCases interface {
	CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error)
	ListCampaigns(ctx context.Context) ([]domain.Campaign, error)
	GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error)
}

type Repository interface {
	CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error)
	ListCampaigns(ctx context.Context) ([]domain.Campaign, error)
	GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error)
}
