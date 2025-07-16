package category

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/usecases/domain"
)

// CategoryRepositoryPort is the repository contract for Category.
type CategoryRepositoryPort interface {
	ListCategories(context.Context) ([]domain.Category, error)
	CreateCategory(context.Context, *domain.Category) (int64, error)
	UpdateCategory(context.Context, *domain.Category) error
	DeleteCategory(context.Context, int64) error
}

type CategoryUseCases struct {
	repo CategoryRepositoryPort
}

func NewCategoryUseCases(repo CategoryRepositoryPort) *CategoryUseCases {
	return &CategoryUseCases{repo: repo}
}

func (u *CategoryUseCases) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return u.repo.ListCategories(ctx)
}
func (u *CategoryUseCases) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	return u.repo.CreateCategory(ctx, c)
}
func (u *CategoryUseCases) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return u.repo.UpdateCategory(ctx, c)
}
func (u *CategoryUseCases) DeleteCategory(ctx context.Context, id int64) error {
	return u.repo.DeleteCategory(ctx, id)
}
