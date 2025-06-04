package lot

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type RepositoryPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLots(context.Context, int64) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	return u.repo.CreateLot(ctx, l)
}

func (u *UseCases) ListLots(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	return u.repo.ListLots(ctx, fieldID)
}

func (u *UseCases) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	return u.repo.GetLot(ctx, id)
}

func (u *UseCases) UpdateLot(ctx context.Context, l *domain.Lot) error {
	return u.repo.UpdateLot(ctx, l)
}

func (u *UseCases) DeleteLot(ctx context.Context, id int64) error {
	return u.repo.DeleteLot(ctx, id)
}
