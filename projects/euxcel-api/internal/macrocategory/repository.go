package macrocategory

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gormAdapter "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/macrocategory/usecases/domain"
)

type repository struct {
	db gormAdapter.Repository
}

// NewRepository creates a new MacroCategory repository.
func NewRepository(db gormAdapter.Repository) Repository {
	return &repository{db: db}
}

func (r *repository) CreateMacroCategory(ctx context.Context, m *domain.MacroCategory) (int64, error) {
	if m == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "macro category is nil", nil)
	}
	model := models.FromDomainMacroCategory(m)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create macro category", err)
	}
	return model.ID, nil
}

func (r *repository) ListMacroCategories(ctx context.Context) ([]domain.MacroCategory, error) {
	var list []models.MacroCategory
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list macro categories", err)
	}
	result := make([]domain.MacroCategory, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, nil
}

func (r *repository) GetMacroCategory(ctx context.Context, id int64) (*domain.MacroCategory, error) {
	var model models.MacroCategory
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("macro category with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get macro category", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateMacroCategory(ctx context.Context, m *domain.MacroCategory) error {
	if m == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "macro category is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.MacroCategory{}).
		Where("id = ?", m.ID).
		Updates(models.FromDomainMacroCategory(m))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update macro category", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("macro category with id %d does not exist", m.ID), nil)
	}
	return nil
}

func (r *repository) DeleteMacroCategory(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).Delete(&models.MacroCategory{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete macro category", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("macro category with id %d does not exist", id), nil)
	}
	return nil
}
