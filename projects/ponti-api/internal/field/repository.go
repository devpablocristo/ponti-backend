// File: internal/field/Repository/Repository.go
package field

import (
	"context"
	"errors"
	"fmt"

	gorm "gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	if f == nil {
		return 0, types.NewError(types.ErrValidation, "field is nil", nil)
	}
	model := models.FromDomain(f)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create field", err)
	}
	return model.ID, nil
}

func (r *Repository) ListFields(ctx context.Context) ([]domain.Field, error) {
	var list []models.Field
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list fields", err)
	}
	result := make([]domain.Field, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, nil
}

// GetField retrieves a field by its ID.
func (r *Repository) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	var model models.Field
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("field with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get field", err)
	}
	return model.ToDomain(), nil
}

// UpdateField updates an existing field.
func (r *Repository) UpdateField(ctx context.Context, f *domain.Field) error {
	if f == nil {
		return types.NewError(types.ErrValidation, "field is nil", nil)
	}
	model := models.FromDomain(f)
	result := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
		Where("id = ?", f.ID).
		Updates(model)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update field", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("field with id %d does not exist", f.ID), nil)
	}
	return nil
}

// DeleteField deletes a field by its ID.
func (r *Repository) DeleteField(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Field{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete field", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("field with id %d does not exist", id), nil)
	}
	return nil
}
