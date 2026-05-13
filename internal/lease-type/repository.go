package leasetype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/lease-type/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
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
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}
	model.Base = sharedmodels.Base{
		CreatedBy: lt.CreatedBy,
		UpdatedBy: lt.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create lease type")
	}
	return model.ID, nil
}

func (r *Repository) ListLeaseTypes(ctx context.Context, page, perPage int) ([]domain.LeaseType, int64, error) {
	var total int64
	base := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.LeaseType{}), "lease_types")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count lease types")
	}

	var list []models.LeaseType
	offset := (page - 1) * perPage
	err := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "lease_types").
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list lease types")
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
	if err := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "lease_types").Where("id = ?", id).First(&model).Error; err != nil {
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
	updateTx := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.LeaseType{}), "lease_types").
		Where("id = ?", lt.ID)
	if !lt.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", lt.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomainLeaseType(lt))
	if result.Error != nil {
		return domainerr.Internal("failed to update lease type")
	}
	if result.RowsAffected == 0 {
		if !lt.UpdatedAt.IsZero() {
			return domainerr.Conflict("lease type not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type with id %d does not exist", lt.ID))
	}
	return nil
}

func (r *Repository) DeleteLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	result := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "lease_types").
		Delete(&models.LeaseType{}, "id = ?", id)
	if result.Error != nil {
		return domainerr.Internal("failed to delete lease type")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type with id %d does not exist", id))
	}
	return nil
}
