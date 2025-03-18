package category

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/category/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository crea una instancia del adaptador GORM para items.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	if c == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "category is nil", nil)
	}
	model := models.FromDomainCategory(c)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create category", err)
	}
	return model.ID, nil
}

func (r *repository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	var list []models.Category
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list categories", err)
	}
	result := make([]domain.Category, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *repository) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	var model models.Category
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("category with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get category", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	if c == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "category is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ?", c.ID).
		Updates(models.FromDomainCategory(c))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update category", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("category with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *repository) DeleteCategory(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Category{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete category", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("category with id %d does not exist", id), nil)
	}
	return nil
}
