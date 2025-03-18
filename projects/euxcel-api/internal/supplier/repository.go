package supplier

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gormAdapter "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/supplier/usecases/domain"
)

type repository struct {
	db gormAdapter.Repository
}

// NewRepository creates a new Supplier repository.
func NewRepository(db gormAdapter.Repository) Repository {
	return &repository{db: db}
}

func (r *repository) CreateSupplier(ctx context.Context, s *domain.Supplier) (int64, error) {
	if s == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "supplier is nil", nil)
	}
	model := models.FromDomainSupplier(s)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create supplier", err)
	}
	return model.ID, nil
}

func (r *repository) ListSuppliers(ctx context.Context) ([]domain.Supplier, error) {
	var list []models.Supplier
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list suppliers", err)
	}
	result := make([]domain.Supplier, 0, len(list))
	for _, s := range list {
		result = append(result, *s.ToDomain())
	}
	return result, nil
}

func (r *repository) GetSupplier(ctx context.Context, id int64) (*domain.Supplier, error) {
	var model models.Supplier
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("supplier with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get supplier", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateSupplier(ctx context.Context, s *domain.Supplier) error {
	if s == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "supplier is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Supplier{}).
		Where("id = ?", s.ID).
		Updates(models.FromDomainSupplier(s))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update supplier", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("supplier with id %d does not exist", s.ID), nil)
	}
	return nil
}

func (r *repository) DeleteSupplier(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).Delete(&models.Supplier{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete supplier", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("supplier with id %d does not exist", id), nil)
	}
	return nil
}
