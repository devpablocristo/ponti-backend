package macrocategory

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory/usecases/domain"
)

type useCases struct {
	repo Repository
}

func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateMacroCategory(ctx context.Context, m *domain.MacroCategory) (int64, error) {
	return u.repo.CreateMacroCategory(ctx, m)
}

func (u *useCases) ListMacroCategories(ctx context.Context) ([]domain.MacroCategory, error) {
	return u.repo.ListMacroCategories(ctx)
}

func (u *useCases) GetMacroCategory(ctx context.Context, id int64) (*domain.MacroCategory, error) {
	return u.repo.GetMacroCategory(ctx, id)
}

func (u *useCases) UpdateMacroCategory(ctx context.Context, m *domain.MacroCategory) error {
	return u.repo.UpdateMacroCategory(ctx, m)
}

func (u *useCases) DeleteMacroCategory(ctx context.Context, id int64) error {
	return u.repo.DeleteMacroCategory(ctx, id)
}
