package investor

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	identity "github.com/devpablocristo/ponti-backend/internal/identity"
	models "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
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

func (r *Repository) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	if err := sharedrepo.ValidateEntity(inv, "investor"); err != nil {
		return 0, err
	}
	model := models.FromDomain(inv)
	model.Base = sharedmodels.Base{
		CreatedBy: inv.CreatedBy,
		UpdatedBy: inv.UpdatedBy,
	}
	create := func(db *gorm.DB) error {
		if err := db.WithContext(ctx).Create(model).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("an investor with that name already exists")
			}
			return domainerr.Internal("failed to create investor")
		}
		// T1.e: dual-write de tenant_id (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			if err := db.WithContext(ctx).Exec("UPDATE investors SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
				return domainerr.Internal("failed to set investor tenant")
			}
		}
		return nil
	}

	// Flag off → comportamiento idéntico al actual.
	if !sharedmodels.IdentityGateEnabled() {
		if err := create(r.db.Client()); err != nil {
			return 0, err
		}
		return model.ID, nil
	}

	// Identity Gate on: investor + resolución de identidad + stamp de actor_id en UNA tx.
	if err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := create(tx); err != nil {
			return err
		}
		res, err := identity.ResolveOrCreateIdentity(ctx, tx, identity.RoleInvestor, identity.ResolveInput{RawName: inv.Name})
		if err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("an entity with that identity already exists")
			}
			return domainerr.Internal("failed to resolve investor identity")
		}
		return tx.Exec("UPDATE investors SET actor_id = ? WHERE id = ?", res.ActorID, model.ID).Error
	}); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *Repository) ListInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	var total int64

	countQ := r.db.Client().WithContext(ctx).Model(&models.Investor{})
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		countQ = countQ.Where("tenant_id = ?", orgID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count investors")
	}

	var list []models.Investor
	offset := (page - 1) * perPage
	listQ := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC")
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		listQ = listQ.Where("tenant_id = ?", orgID)
	}
	err := listQ.Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list investors")
	}

	result := make([]domain.Investor, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	var model models.Investor
	q := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — 404 si el investor no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "investor", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	if err := sharedrepo.ValidateEntity(inv, "investor"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(inv.ID, "investor"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Investor{}).
		Where("id = ?", inv.ID)
	if !inv.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", inv.UpdatedAt)
	}
	// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
	}
	result := updateTx.Updates(models.FromDomain(inv))
	if result.Error != nil {
		return domainerr.Internal("failed to update investor")
	}
	if result.RowsAffected == 0 {
		if !inv.UpdatedAt.IsZero() {
			return domainerr.Conflict("investor not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", inv.ID))
	}
	return nil
}

func (r *Repository) DeleteInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	delTx := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — solo borra si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		delTx = delTx.Where("tenant_id = ?", orgID)
	}
	result := delTx.Delete(&models.Investor{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete investor")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", id))
	}
	return nil
}

func (r *Repository) ArchiveInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	archiveTx := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — solo archiva si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		archiveTx = archiveTx.Where("tenant_id = ?", orgID)
	}
	result := archiveTx.Delete(&models.Investor{})
	if result.Error != nil {
		return domainerr.Internal("failed to archive investor")
	}
	// Idempotente: si ya estaba archivado, RowsAffected == 0 es OK
	return nil
}

func (r *Repository) RestoreInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	restoreTx := r.db.Client().WithContext(ctx).Unscoped().
		Model(&models.Investor{}).
		Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — solo restaura si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		restoreTx = restoreTx.Where("tenant_id = ?", orgID)
	}
	result := restoreTx.Update("deleted_at", nil)
	if result.Error != nil {
		return domainerr.Internal("failed to restore investor")
	}
	return nil
}
