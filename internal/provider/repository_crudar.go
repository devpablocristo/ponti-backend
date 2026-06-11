package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	identity "github.com/devpablocristo/ponti-backend/internal/identity"
	models "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
)

// CreateProvider crea un proveedor. Con el Identity Gate on resuelve la identidad
// (rol provider) y estampa actor_id en la misma tx (mismo patrón que CreateCustomer).
func (r *Repository) CreateProvider(ctx context.Context, p *domain.Provider) (int64, error) {
	if err := sharedrepo.ValidateEntity(p, "provider"); err != nil {
		return 0, err
	}
	model := models.FromDomain(p)
	model.Base = sharedmodels.Base{CreatedBy: p.CreatedBy, UpdatedBy: p.UpdatedBy}

	create := func(db *gorm.DB) error {
		if err := db.WithContext(ctx).Create(model).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a provider with that name already exists")
			}
			return domainerr.Internal("failed to create provider")
		}
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			if err := db.WithContext(ctx).Exec("UPDATE providers SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
				return domainerr.Internal("failed to set provider tenant")
			}
		}
		return nil
	}

	if !sharedmodels.IdentityGateEnabled() {
		if err := create(r.db.Client()); err != nil {
			return 0, err
		}
		return model.ID, nil
	}

	if err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := create(tx); err != nil {
			return err
		}
		res, err := identity.ResolveOrCreateIdentity(ctx, tx, identity.RoleProvider, identity.ResolveInput{RawName: p.Name, TaxID: p.TaxID})
		if err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("an entity with that identity already exists")
			}
			return domainerr.Internal("failed to resolve provider identity")
		}
		return tx.Exec("UPDATE providers SET actor_id = ? WHERE id = ?", res.ActorID, model.ID).Error
	}); err != nil {
		return 0, err
	}
	return model.ID, nil
}

// GetArchivedProviders lista los proveedores archivados del tenant.
func (r *Repository) GetArchivedProviders(ctx context.Context) ([]domain.Provider, error) {
	var providers []models.Provider
	db0 := r.db.Client().WithContext(ctx).Unscoped().
		Model(&models.Provider{}).
		Where("deleted_at IS NOT NULL")
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		db0 = db0.Where("tenant_id = ?", orgID)
	}
	if err := db0.Find(&providers).Error; err != nil {
		return nil, domainerr.Internal("failed to list archived providers")
	}
	res := make([]domain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}

// GetProvider devuelve un proveedor activo del tenant (404 si no es del tenant).
func (r *Repository) GetProvider(ctx context.Context, id int64) (*domain.Provider, error) {
	var model models.Provider
	q := r.db.Client().WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider with id %d not found", id))
		}
		return nil, domainerr.Internal("failed to get provider")
	}
	return model.ToDomain(), nil
}

// UpdateProvider renombra un proveedor (guard de tenant + dedup vía índice/trigger).
func (r *Repository) UpdateProvider(ctx context.Context, p *domain.Provider) error {
	if err := sharedrepo.ValidateEntity(p, "provider"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(p.ID, "provider"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Provider{}).
		Where("id = ?", p.ID)
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
	}
	result := updateTx.Updates(map[string]any{"name": p.Name, "updated_by": p.UpdatedBy})
	if result.Error != nil {
		if sharedrepo.IsUniqueViolation(result.Error) {
			return domainerr.Conflict("a provider with that name already exists")
		}
		return domainerr.Internal("failed to update provider")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider with id %d does not exist", p.ID))
	}
	return nil
}

// DeleteProvider hard-borra un proveedor del tenant.
func (r *Repository) DeleteProvider(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return err
	}
	deleteTx := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id)
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		deleteTx = deleteTx.Where("tenant_id = ?", orgID)
	}
	result := deleteTx.Delete(&models.Provider{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete provider")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider with id %d does not exist", id))
	}
	return nil
}

// ArchiveProvider soft-borra un proveedor del tenant.
func (r *Repository) ArchiveProvider(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p models.Provider
		loadQ := tx.Unscoped().Where("id = ?", id)
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider %d not found", id))
			}
			return domainerr.Internal("failed to get provider")
		}
		if p.DeletedAt.Valid {
			return domainerr.Conflict("provider already archived")
		}
		return tx.Model(&models.Provider{}).Where("id = ?", id).
			Updates(map[string]any{"deleted_at": time.Now(), "deleted_by": gorm.Expr("NULL")}).Error
	})
}

// RestoreProvider reactiva un proveedor archivado (el dedup puede rechazar → 409).
func (r *Repository) RestoreProvider(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "provider"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p models.Provider
		loadQ := tx.Unscoped().Where("id = ?", id)
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("provider %d not found", id))
			}
			return domainerr.Internal("failed to get provider")
		}
		if !p.DeletedAt.Valid {
			return domainerr.Conflict("provider is not archived")
		}
		if err := tx.Unscoped().Model(&models.Provider{}).Where("id = ?", id).
			Updates(map[string]any{"deleted_at": nil, "deleted_by": nil, "updated_at": time.Now()}).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a provider with that name already exists; cannot restore")
			}
			return domainerr.Internal("failed to restore provider")
		}
		return nil
	})
}
