package leasetype

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/lease-type/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/lease-type/usecases/domain"
	"github.com/devpablocristo/platform/persistence/gorm/go/tenancy"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
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
	base := tenancy.Scope(ctx, r.db.Client().WithContext(ctx).Model(&models.LeaseType{}), "lease_types")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count lease types")
	}

	var list []models.LeaseType
	offset := (page - 1) * perPage
	err := tenancy.Scope(ctx, r.db.Client().WithContext(ctx), "lease_types").
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
	if err := tenancy.Scope(ctx, r.db.Client().WithContext(ctx), "lease_types").Where("id = ?", id).First(&model).Error; err != nil {
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
	updateTx := tenancy.Scope(ctx, r.db.Client().WithContext(ctx).Model(&models.LeaseType{}), "lease_types").
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

func (r *Repository) ListArchivedLeaseTypes(ctx context.Context, page, perPage int) ([]domain.LeaseType, int64, error) {
	var total int64
	base := tenancy.Scope(ctx, r.db.Client().WithContext(ctx).Unscoped().Model(&models.LeaseType{}), "lease_types").
		Where("deleted_at IS NOT NULL")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived lease types")
	}
	var list []models.LeaseType
	if err := base.Offset((page - 1) * perPage).Limit(perPage).Order("deleted_at DESC").Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived lease types")
	}
	out := make([]domain.LeaseType, 0, len(list))
	for i := range list {
		out = append(out, *list[i].ToDomain())
	}
	return out, total, nil
}

func (r *Repository) ArchiveLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.LeaseType
		if err := tenancy.Scope(ctx, tx.Unscoped(), "lease_types").Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type %d not found", id))
			}
			return domainerr.Internal("failed to get lease type")
		}
		if item.DeletedAt.Valid {
			return domainerr.Conflict("lease type already archived")
		}
		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, item.TenantID, "lease_types", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := tenancy.Scope(ctx, tx.Model(&models.LeaseType{}), "lease_types").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "lease_types", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive lease type")
		}
		return nil
	})
}

func (r *Repository) RestoreLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.LeaseType
		if err := tenancy.Scope(ctx, tx.Unscoped(), "lease_types").Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type %d not found", id))
			}
			return domainerr.Internal("failed to get lease type")
		}
		if !item.DeletedAt.Valid {
			return domainerr.Conflict("lease type is not archived")
		}
		if err := tenancy.Scope(ctx, tx.Unscoped().Model(&models.LeaseType{}), "lease_types").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "lease_types", time.Now())).Error; err != nil {
			return domainerr.Internal("failed to restore lease type")
		}
		return nil
	})
}

func (r *Repository) HardDeleteLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		leaseDB := tenancy.Scope(ctx, tx.Unscoped().Table("lease_types"), "lease_types")
		var count int64
		if err := leaseDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check lease type existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type with id %d does not exist", id))
		}
		if err := lifecycle.RequireArchived(leaseDB, "lease_types", "lease type", id); err != nil {
			return err
		}
		var fields int64
		if err := tenancy.Scope(ctx, tx.Unscoped().Table("fields"), "fields").Where("lease_type_id = ?", id).Count(&fields).Error; err != nil {
			return domainerr.Internal("failed to check fields")
		}
		if fields > 0 {
			return domainerr.Conflict(fmt.Sprintf("lease type has %d field reference(s); remove them first", fields))
		}
		if err := tenancy.Scope(ctx, tx.Unscoped(), "lease_types").Delete(&models.LeaseType{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete lease type")
		}
		return nil
	})
}
