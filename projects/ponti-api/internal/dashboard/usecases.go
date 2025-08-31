package dashboard

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type RepositoryPort interface {
	CreateDashboard(context.Context, *domain.Dashboard) (int64, error)
	ListDashboards(context.Context) ([]domain.Dashboard, error)
	GetDashboard(context.Context, int64) (*domain.Dashboard, error)
	UpdateDashboard(context.Context, *domain.Dashboard) error
	DeleteDashboard(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases creates a new instance of Dashboard use cases.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateDashboard(ctx context.Context, d *domain.Dashboard) (int64, error) {
	return u.repo.CreateDashboard(ctx, d)
}

func (u *UseCases) ListDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	return u.repo.ListDashboards(ctx)
}

func (u *UseCases) GetDashboard(ctx context.Context, id int64) (*domain.Dashboard, error) {
	return u.repo.GetDashboard(ctx, id)
}

func (u *UseCases) UpdateDashboard(ctx context.Context, d *domain.Dashboard) error {
	return u.repo.UpdateDashboard(ctx, d)
}

func (u *UseCases) DeleteDashboard(ctx context.Context, id int64) error {
	return u.repo.DeleteDashboard(ctx, id)
}
