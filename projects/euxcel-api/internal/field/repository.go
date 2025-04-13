package field

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository creates a new GORM repository instance for Field.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	if f == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "field is nil", nil)
	}
	model := models.FromDomainField(f)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create field", err)
	}
	return model.ID, nil
}

func (r *repository) ListFields(ctx context.Context) ([]domain.Field, error) {
	var list []models.Field
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list fields", err)
	}
	result := make([]domain.Field, 0, len(list))
	for _, f := range list {
		result = append(result, *f.ToDomain())
	}
	return result, nil
}

func (r *repository) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	var model models.Field
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("field with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get field", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateField(ctx context.Context, f *domain.Field) error {
	if f == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "field is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
		Where("id = ?", f.ID).
		Updates(models.FromDomainField(f))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update field", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("field with id %d does not exist", f.ID), nil)
	}
	return nil
}

func (r *repository) DeleteField(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Field{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete field", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("field with id %d does not exist", id), nil)
	}
	return nil
}
