package supply

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type RepositoryPort interface {
	CreateSupply(context.Context, *domain.Supply) (int64, error)
	GetSupply(context.Context, int64) (*domain.Supply, error)
	UpdateSupply(context.Context, *domain.Supply) error
	DeleteSupply(context.Context, int64) error
	ListSupplies(context.Context, SupplyFilters) ([]domain.Supply, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateSupply(ctx context.Context, s *domain.Supply) (int64, error) {
	return u.repo.CreateSupply(ctx, s)
}
func (u *UseCases) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	return u.repo.GetSupply(ctx, id)
}
func (u *UseCases) UpdateSupply(ctx context.Context, s *domain.Supply) error {
	return u.repo.UpdateSupply(ctx, s)
}
func (u *UseCases) DeleteSupply(ctx context.Context, id int64) error {
	return u.repo.DeleteSupply(ctx, id)
}
func (u *UseCases) ListSupplies(ctx context.Context, filters SupplyFilters) ([]domain.Supply, error) {
	return u.repo.ListSupplies(ctx, filters)
}
