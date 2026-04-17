package category

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/internal/category/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/category/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateCategory(ctx context.Context, c *domain.Category) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "category"); err != nil {
		return 0, err
	}
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create category", err)
	}
	return model.ID, nil
}

func (r *Repository) ListCategories(ctx context.Context, page, perPage int) ([]domain.Category, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).Model(&models.Category{}).Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count categories", err)
	}

	var list []models.Category
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list categories", err)
	}

	result := make([]domain.Category, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	if err := sharedrepo.ValidateID(id, "category"); err != nil {
		return nil, err
	}
	var model models.Category
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "category", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCategory(ctx context.Context, c *domain.Category) error {
	if err := sharedrepo.ValidateEntity(c, "category"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "category"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomain(c))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update category", result.Error)
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "category not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, fmt.Sprintf("category with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "category"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Category{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete category", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("category with id %d does not exist", id), nil)
	}
	return nil
}
