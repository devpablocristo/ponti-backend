package category

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/internal/category/usecases/domain"
)

type RepositoryPort interface {
	CreateCategory(context.Context, *domain.Category) (int64, error)
	ListCategories(context.Context, int, int) ([]domain.Category, int64, error)
	GetCategory(context.Context, int64) (*domain.Category, error)
	UpdateCategory(context.Context, *domain.Category) error
	DeleteCategory(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	return u.repo.CreateCategory(ctx, c)
}

func (u *UseCases) ListCategories(ctx context.Context, page, perPage int) ([]domain.Category, int64, error) {
	return u.repo.ListCategories(ctx, page, perPage)
}

func (u *UseCases) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	return u.repo.GetCategory(ctx, id)
}

func (u *UseCases) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return u.repo.UpdateCategory(ctx, c)
}

func (u *UseCases) DeleteCategory(ctx context.Context, id int64) error {
	return u.repo.DeleteCategory(ctx, id)
}
