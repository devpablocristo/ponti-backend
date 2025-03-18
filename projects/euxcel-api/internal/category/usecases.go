package category

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category/usecases/domain"
)

type useCases struct {
	repo Repository
}

func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	return u.repo.CreateCategory(ctx, c)
}

func (u *useCases) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return u.repo.ListCategories(ctx)
}

func (u *useCases) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	return u.repo.GetCategory(ctx, id)
}

func (u *useCases) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return u.repo.UpdateCategory(ctx, c)
}

func (u *useCases) DeleteCategory(ctx context.Context, id int64) error {
	return u.repo.DeleteCategory(ctx, id)
}
