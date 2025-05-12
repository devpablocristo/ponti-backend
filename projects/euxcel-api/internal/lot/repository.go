package lot

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository crea una nueva instancia de repositorio para Lot.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	if l == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "lot is nil", nil)
	}
	model := models.FromDomainLot(l)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create lot", err)
	}
	// model.ID ya es int64
	return model.ID, nil
}

func (r *repository) ListLots(ctx context.Context) ([]domain.Lot, error) {
	var list []models.Lot
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list lots", err)
	}
	result := make([]domain.Lot, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, nil
}

func (r *repository) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	var model models.Lot
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("lot with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get lot", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateLot(ctx context.Context, l *domain.Lot) error {
	if l == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "lot is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Lot{}).
		Where("id = ?", l.ID).
		Updates(models.FromDomainLot(l))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update lot", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("lot with id %d does not exist", l.ID), nil)
	}
	return nil
}

func (r *repository) DeleteLot(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Lot{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete lot", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("lot with id %d does not exist", id), nil)
	}
	return nil
}
