package campaign

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type useCases struct {
	repo Repository
}

func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error) {
	return u.repo.CreateCampaign(ctx, c)
}

func (u *useCases) ListCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	return u.repo.ListCampaigns(ctx)
}

func (u *useCases) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	return u.repo.GetCampaign(ctx, id)
}
