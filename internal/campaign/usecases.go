package campaign

import (
	"context"

	"github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
)

type RepositoryPort interface {
	CreateCampaign(context.Context, *domain.Campaign) (int64, error)
	ListCampaigns(context.Context, int64, string) ([]domain.Campaign, error)
	GetCampaign(context.Context, int64) (*domain.Campaign, error)
	GetArchivedCampaigns(context.Context) ([]domain.Campaign, error)
	UpdateCampaign(context.Context, *domain.Campaign) error
	DeleteCampaign(context.Context, int64) error
	ArchiveCampaign(context.Context, int64) error
	RestoreCampaign(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error) {
	return u.repo.CreateCampaign(ctx, c)
}

func (u *UseCases) ListCampaigns(ctx context.Context, customerID int64, projectName string) ([]domain.Campaign, error) {
	return u.repo.ListCampaigns(ctx, customerID, projectName)
}

func (u *UseCases) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	return u.repo.GetCampaign(ctx, id)
}

func (u *UseCases) GetArchivedCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	return u.repo.GetArchivedCampaigns(ctx)
}

func (u *UseCases) UpdateCampaign(ctx context.Context, c *domain.Campaign) error {
	return u.repo.UpdateCampaign(ctx, c)
}

func (u *UseCases) DeleteCampaign(ctx context.Context, id int64) error {
	return u.repo.DeleteCampaign(ctx, id)
}

func (u *UseCases) ArchiveCampaign(ctx context.Context, id int64) error {
	return u.repo.ArchiveCampaign(ctx, id)
}

func (u *UseCases) RestoreCampaign(ctx context.Context, id int64) error {
	return u.repo.RestoreCampaign(ctx, id)
}
