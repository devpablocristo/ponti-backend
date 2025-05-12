package lot

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

type useCases struct {
	repo Repository
}

func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

// CreateLot creates a lot and returns its new ID.
func (u *useCases) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	return u.repo.CreateLot(ctx, l)
}

// ListLots returns all lots.
func (u *useCases) ListLots(ctx context.Context) ([]domain.Lot, error) {
	return u.repo.ListLots(ctx)
}

// GetLot retrieves a lot by ID.
func (u *useCases) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	return u.repo.GetLot(ctx, id)
}

// UpdateLot updates an existing lot.
func (u *useCases) UpdateLot(ctx context.Context, l *domain.Lot) error {
	return u.repo.UpdateLot(ctx, l)
}

// DeleteLot deletes a lot by ID.
func (u *useCases) DeleteLot(ctx context.Context, id int64) error {
	return u.repo.DeleteLot(ctx, id)
}
