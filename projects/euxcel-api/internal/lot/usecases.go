package lot

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases creates a new instance of Lot use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	return u.repo.CreateLot(ctx, l)
}

func (u *useCases) ListLots(ctx context.Context) ([]domain.Lot, error) {
	return u.repo.ListLots(ctx)
}

func (u *useCases) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	return u.repo.GetLot(ctx, id)
}

func (u *useCases) UpdateLot(ctx context.Context, l *domain.Lot) error {
	return u.repo.UpdateLot(ctx, l)
}

func (u *useCases) DeleteLot(ctx context.Context, id int64) error {
	return u.repo.DeleteLot(ctx, id)
}
