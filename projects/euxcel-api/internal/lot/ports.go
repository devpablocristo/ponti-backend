package lot

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// UseCases define las operaciones de negocio para Lot.
type UseCases interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLots(context.Context) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
}

// Repository define las operaciones de persistencia para Lot.
type Repository interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLots(context.Context) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
}
