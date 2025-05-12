package crop

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop/usecases/domain"
)

// UseCases defines business operations for Crop.
type UseCases interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context) ([]domain.Crop, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	DeleteCrop(context.Context, int64) error
}

// Repository defines data persistence operations for Crop.
type Repository interface {
	CreateCrop(context.Context, *domain.Crop) (int64, error)
	ListCrops(context.Context) ([]domain.Crop, error)
	GetCrop(context.Context, int64) (*domain.Crop, error)
	UpdateCrop(context.Context, *domain.Crop) error
	DeleteCrop(context.Context, int64) error
}
