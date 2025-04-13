package crop

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop/usecases/domain"
)

// UseCases defines business operations for Crop.
type UseCases interface {
	CreateCrop(ctx context.Context, c *domain.Crop) (int64, error)
	ListCrops(ctx context.Context) ([]domain.Crop, error)
	GetCrop(ctx context.Context, id int64) (*domain.Crop, error)
	UpdateCrop(ctx context.Context, c *domain.Crop) error
	DeleteCrop(ctx context.Context, id int64) error
}

// Repository defines data persistence operations for Crop.
type Repository interface {
	CreateCrop(ctx context.Context, c *domain.Crop) (int64, error)
	ListCrops(ctx context.Context) ([]domain.Crop, error)
	GetCrop(ctx context.Context, id int64) (*domain.Crop, error)
	UpdateCrop(ctx context.Context, c *domain.Crop) error
	DeleteCrop(ctx context.Context, id int64) error
}
