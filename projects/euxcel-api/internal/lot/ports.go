package lot

import (
	"context"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

type UseCases interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLots(context.Context) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
}

type Repository interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLots(context.Context) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
}
