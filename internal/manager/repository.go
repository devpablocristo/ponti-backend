package manager

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/manager/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
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

func (r *Repository) CreateManager(ctx context.Context, m *domain.Manager) (int64, error) {
	if err := sharedrepo.ValidateEntity(m, "manager"); err != nil {
		return 0, err
	}
	model := models.FromDomain(m)
	model.Base = sharedmodels.Base{
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create manager")
	}
	// T1.e: dual-write de tenant_id (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE managers SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set manager tenant")
		}
	}
	return model.ID, nil
}

func (r *Repository) ListManagers(ctx context.Context, page, perPage int) ([]domain.Manager, int64, error) {
	var total int64
	countTx := r.db.Client().WithContext(ctx).Model(&models.Manager{})
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		countTx = countTx.Where("tenant_id = ?", orgID)
	}
	if err := countTx.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count managers")
	}

	var list []models.Manager
	offset := (page - 1) * perPage
	listTx := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC")
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		listTx = listTx.Where("tenant_id = ?", orgID)
	}
	err := listTx.Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list managers")
	}

	result := make([]domain.Manager, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return nil, err
	}
	var model models.Manager
	q := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — 404 si el manager no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "manager", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateManager(ctx context.Context, m *domain.Manager) error {
	if err := sharedrepo.ValidateEntity(m, "manager"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(m.ID, "manager"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Manager{}).
		Where("id = ?", m.ID)
	if !m.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", m.UpdatedAt)
	}
	// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
	}
	result := updateTx.Updates(models.FromDomain(m))
	if result.Error != nil {
		return domainerr.Internal("failed to update manager")
	}
	if result.RowsAffected == 0 {
		if !m.UpdatedAt.IsZero() {
			return domainerr.Conflict("manager not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", m.ID))
	}
	return nil
}

// DeleteManager elimina físicamente (hard delete) un manager.
func (r *Repository) DeleteManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	deleteTx := r.db.Client().WithContext(ctx).
		Unscoped().
		Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		deleteTx = deleteTx.Where("tenant_id = ?", orgID)
	}
	result := deleteTx.Delete(&models.Manager{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete manager")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", id))
	}
	return nil
}

// ArchiveManager archiva (soft delete) un manager. Idempotente.
func (r *Repository) ArchiveManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	archiveTx := r.db.Client().WithContext(ctx).
		Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		archiveTx = archiveTx.Where("tenant_id = ?", orgID)
	}
	result := archiveTx.Delete(&models.Manager{})
	if result.Error != nil {
		return domainerr.Internal("failed to archive manager")
	}
	return nil
}

// RestoreManager restaura un manager archivado.
func (r *Repository) RestoreManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	restoreTx := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Manager{}).
		Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		restoreTx = restoreTx.Where("tenant_id = ?", orgID)
	}
	result := restoreTx.Update("deleted_at", nil)
	if result.Error != nil {
		return domainerr.Internal("failed to restore manager")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", id))
	}
	return nil
}
