package lot

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// UseCases defines business operations for Lot.
type UseCases interface {
	CreateLot(ctx context.Context, l *domain.Lot) (int64, error)
	ListLots(ctx context.Context) ([]domain.Lot, error)
	GetLot(ctx context.Context, id int64) (*domain.Lot, error)
	UpdateLot(ctx context.Context, l *domain.Lot) error
	DeleteLot(ctx context.Context, id int64) error
}

// Repository defines data persistence operations for Lot.
type Repository interface {
	CreateLot(ctx context.Context, l *domain.Lot) (int64, error)
	ListLots(ctx context.Context) ([]domain.Lot, error)
	GetLot(ctx context.Context, id int64) (*domain.Lot, error)
	UpdateLot(ctx context.Context, l *domain.Lot) error
	DeleteLot(ctx context.Context, id int64) error
}
