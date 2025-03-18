package macrocategory

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory/usecases/domain"
)

// UseCases defines business operations for MacroCategory.
type UseCases interface {
	CreateMacroCategory(ctx context.Context, m *domain.MacroCategory) (int64, error)
	ListMacroCategories(ctx context.Context) ([]domain.MacroCategory, error)
	GetMacroCategory(ctx context.Context, id int64) (*domain.MacroCategory, error)
	UpdateMacroCategory(ctx context.Context, m *domain.MacroCategory) error
	DeleteMacroCategory(ctx context.Context, id int64) error
}

// Repository defines persistence operations for MacroCategory.
type Repository interface {
	CreateMacroCategory(ctx context.Context, m *domain.MacroCategory) (int64, error)
	ListMacroCategories(ctx context.Context) ([]domain.MacroCategory, error)
	GetMacroCategory(ctx context.Context, id int64) (*domain.MacroCategory, error)
	UpdateMacroCategory(ctx context.Context, m *domain.MacroCategory) error
	DeleteMacroCategory(ctx context.Context, id int64) error
}
