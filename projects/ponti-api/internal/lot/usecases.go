package lot

import (
	"context"
	"fmt"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
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
	crop crop.UseCasesPort
}

func NewUseCases(repo RepositoryPort, crop crop.UseCasesPort) *UseCases {
	return &UseCases{
		repo: repo,
		crop: crop,
	}
}

func (u *UseCases) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	return u.repo.CreateLot(ctx, l)
}

func (u *UseCases) ListLots(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	lots, err := u.repo.ListLots(ctx, fieldID)
	if err != nil {
		return nil, err
	}
	for i := range lots {
		if err := u.enrichLot(ctx, &lots[i]); err != nil {
			return nil, err
		}
	}
	return lots, nil
}

func (u *UseCases) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	l, err := u.repo.GetLot(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.enrichLot(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (u *UseCases) UpdateLot(ctx context.Context, l *domain.Lot) error {
	return u.repo.UpdateLot(ctx, l)
}

func (u *UseCases) DeleteLot(ctx context.Context, id int64) error {
	return u.repo.DeleteLot(ctx, id)
}

// helpers
func (u *UseCases) enrichLot(ctx context.Context, l *domain.Lot) error {
	prev, err := u.crop.GetCrop(ctx, l.PreviousCrop.ID)
	if err != nil {
		return fmt.Errorf("fetch previous crop %d: %w", l.PreviousCrop.ID, err)
	}
	l.PreviousCrop = *prev

	// Cargar CurrentCrop
	cur, err := u.crop.GetCrop(ctx, l.CurrentCrop.ID)
	if err != nil {
		return fmt.Errorf("fetch current crop %d: %w", l.CurrentCrop.ID, err)
	}
	l.CurrentCrop = *cur

	return nil
}
