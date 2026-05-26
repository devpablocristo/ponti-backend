package category

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
)

type RepositoryPort interface {
	CreateCategory(context.Context, *domain.Category) (int64, error)
	ListCategories(ctx context.Context, filters domain.ListFilters, page, perPage int) ([]domain.Category, int64, error)
	ListArchivedCategories(context.Context, int, int) ([]domain.Category, int64, error)
	GetCategory(context.Context, int64) (*domain.Category, error)
	UpdateCategory(context.Context, *domain.Category) error
	ArchiveCategory(context.Context, int64) error
	RestoreCategory(context.Context, int64) error
	HardDeleteCategory(context.Context, int64) error
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

func (u *UseCases) ListCategories(ctx context.Context, filters domain.ListFilters, page, perPage int) ([]domain.Category, int64, error) {
	return u.repo.ListCategories(ctx, filters, page, perPage)
}

func (u *UseCases) ListArchivedCategories(ctx context.Context, page, perPage int) ([]domain.Category, int64, error) {
	return u.repo.ListArchivedCategories(ctx, page, perPage)
}

func (u *UseCases) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	return u.repo.GetCategory(ctx, id)
}

func (u *UseCases) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return u.repo.UpdateCategory(ctx, c)
}

func (u *UseCases) ArchiveCategory(ctx context.Context, id int64) error {
	return u.repo.ArchiveCategory(ctx, id)
}

func (u *UseCases) RestoreCategory(ctx context.Context, id int64) error {
	return u.repo.RestoreCategory(ctx, id)
}

func (u *UseCases) HardDeleteCategory(ctx context.Context, id int64) error {
	return u.repo.HardDeleteCategory(ctx, id)
}
