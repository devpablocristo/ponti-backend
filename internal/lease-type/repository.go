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
	model.Base = sharedmodels.Base{
		CreatedBy: lt.CreatedBy,
		UpdatedBy: lt.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		if sharedrepo.IsUniqueViolation(err) {
			return 0, domainerr.Conflict("a lease type with that name already exists")
		}
		return 0, domainerr.Internal("failed to create lease type")
	}
	// T3 (Modelo 2): dual-write de tenant_id del tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE lease_types SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set lease type tenant")
		}
	}
	return model.ID, nil
}

func (r *Repository) ListLeaseTypes(ctx context.Context, page, perPage int) ([]domain.LeaseType, int64, error) {
	// T3 (Modelo 2): acotar al tenant activo (flag-gated).
	orgID, tenantScoped := sharedmodels.OrgIDFromContext(ctx)
	tenantScoped = tenantScoped && sharedmodels.TenantEnforcementEnabled()

	var total int64
	countTx := r.db.Client().WithContext(ctx).Model(&models.LeaseType{})
	if tenantScoped {
		countTx = countTx.Where("tenant_id = ?", orgID)
	}
	if err := countTx.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count lease types")
	}

	var list []models.LeaseType
	offset := (page - 1) * perPage
	listTx := r.db.Client().WithContext(ctx)
	if tenantScoped {
		listTx = listTx.Where("tenant_id = ?", orgID)
	}
	err := listTx.
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
	q := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T3 (Modelo 2): guard de ownership (flag-gated) — NotFound si no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
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
	// T3 (Modelo 2): guard de ownership (flag-gated) — solo actualiza si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
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

func (r *Repository) ArchiveLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var leaseType models.LeaseType
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T3 (Modelo 2): guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&leaseType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type %d not found", id))
			}
			return domainerr.Internal("failed to get lease type")
		}
		if leaseType.DeletedAt.Valid {
			return domainerr.Conflict("lease type already archived")
		}

		updates := map[string]any{
			"deleted_at": time.Now(),
		}
		updates["deleted_by"] = gorm.Expr("NULL")

		if err := tx.Model(&models.LeaseType{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
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
		var leaseType models.LeaseType
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T3 (Modelo 2): guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&leaseType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type %d not found", id))
			}
			return domainerr.Internal("failed to get lease type")
		}
		if !leaseType.DeletedAt.Valid {
			return domainerr.Conflict("lease type is not archived")
		}

		// La reactivación dispara el trigger de dedup normalize_name; un unique
		// violation se mapea a 409 (no se puede restaurar por nombre duplicado).
		if err := tx.Unscoped().Model(&models.LeaseType{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a lease type with that name already exists; cannot restore")
			}
			return domainerr.Internal("failed to restore lease type")
		}
		return nil
	})
}

func (r *Repository) DeleteLeaseType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lease type"); err != nil {
		return err
	}
	deleteTx := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T3 (Modelo 2): guard de ownership (flag-gated) — NotFound si no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		deleteTx = deleteTx.Where("tenant_id = ?", orgID)
	}
	result := deleteTx.Delete(&models.LeaseType{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete lease type")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lease type with id %d does not exist", id))
	}
	return nil
}
