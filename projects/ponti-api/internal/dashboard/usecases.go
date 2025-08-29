package dashboard

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type RepositoryPort interface {
	GetDashboard(context.Context, domain.DashboardFilter) (*domain.DashboardResponse, error)
	GetDashboardPayload(context.Context, domain.DashboardFilter) ([]byte, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (uc *UseCases) GetDashboard(ctx context.Context, filt domain.DashboardFilter) (*domain.DashboardResponse, error) {
	return uc.repo.GetDashboard(ctx, filt)
}

func (uc *UseCases) GetDashboardPayload(ctx context.Context, filt domain.DashboardFilter) ([]byte, error) {
	return uc.repo.GetDashboardPayload(ctx, filt)
}
