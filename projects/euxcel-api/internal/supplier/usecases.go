package supplier

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/usecases/domain"
)

type useCases struct {
	repo Repository
}

func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateSupplier(ctx context.Context, s *domain.Supplier) (int64, error) {
	return u.repo.CreateSupplier(ctx, s)
}

func (u *useCases) ListSuppliers(ctx context.Context) ([]domain.Supplier, error) {
	return u.repo.ListSuppliers(ctx)
}

func (u *useCases) GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error) {
	return u.repo.GetSupplier(ctx, id)
}

func (u *useCases) UpdateSupplier(ctx context.Context, s *domain.Supplier) error {
	return u.repo.UpdateSupplier(ctx, s)
}

func (u *useCases) DeleteSupplier(ctx context.Context, id int64) error {
	return u.repo.DeleteSupplier(ctx, id)
}
