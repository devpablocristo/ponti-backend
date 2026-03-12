package leasetype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/internal/lease-type/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/lease-type/usecases/domain"
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

func (r *Repository) CreateLeaseType(ctx context.Context, lt *domain.LeaseType) (int64, error) {
	if err := sharedrepo.ValidateEntity(lt, "lease type"); err != nil {
		return 0, err
	}
	model := models.FromDomainLeaseType(lt)
	model.Base = sharedmodels.Base{
		CreatedBy: lt.CreatedBy,
		UpdatedBy: lt.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create lease type", err)
	}
	return model.ID, nil
}

func (r *Repository) ListLeaseTypes(ctx context.Context, page, perPage int) ([]domain.LeaseType, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).Model(&models.LeaseType{}).Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count lease types", err)
	}

	var list []models.LeaseType
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list lease types", err)
	}

	result := make([]domain.LeaseType, 0, len(list))
	for _, lt := range list {
		result = append(result, *lt.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetLeaseType(ctx context.Context, id int64) (*domain.LeaseType, error) {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return nil, err
	}
	var model models.LeaseType
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "lease type", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateLeaseType(ctx context.Context, lt *domain.LeaseType) error {
	if err := sharedrepo.ValidateEntity(lt, "lease type"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(lt.ID, "lease type"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.LeaseType{}).
		Where("id = ?", lt.ID)
	if !lt.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", lt.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomainLeaseType(lt))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update lease type", result.Error)
	}
	if result.RowsAffected == 0 {
		if !lt.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "lease type not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, fmt.Sprintf("lease type with id %d does not exist", lt.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Delete(&models.LeaseType{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete lease type", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("lease type with id %d does not exist", id), nil)
	}
	return nil
}
