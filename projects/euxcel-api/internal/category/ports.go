package category

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category/usecases/domain"
)

// UseCases defines business operations for Category.
type UseCases interface {
	CreateCategory(ctx context.Context, c *domain.Category) (int64, error)
	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategory(ctx context.Context, id int64) (*domain.Category, error)
	UpdateCategory(ctx context.Context, c *domain.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}

// Repository defines operations for Category.
type Repository interface {
	CreateCategory(ctx context.Context, c *domain.Category) (int64, error)
	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategory(ctx context.Context, id int64) (*domain.Category, error)
	UpdateCategory(ctx context.Context, c *domain.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}
