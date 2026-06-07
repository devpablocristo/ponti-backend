package bparams

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/business-parameters/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
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

func (r *Repository) GetByKey(ctx context.Context, key string) (*domain.BusinessParameter, error) {
	var m models.BusinessParameter
	q := r.db.Client().WithContext(ctx).
		Where("key = ?", key)
	// T1.e: guard de ownership por identidad (flag-gated) — NotFound si la key no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	err := q.First(&m).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerr.NotFound("business parameter not found")
	}
	if err != nil {
		return nil, domainerr.Internal("failed to get business parameter")
	}
	return m.ToDomain(), nil
}

func (r *Repository) ListByCategory(ctx context.Context, category string) ([]domain.BusinessParameter, error) {
	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.BusinessParameter{}).
		Where("category = ?", category)

	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		tx = tx.Where("tenant_id = ?", orgID)
	}

	var rows []models.BusinessParameter
	if err := tx.Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list business parameters by category")
	}

	out := make([]domain.BusinessParameter, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]domain.BusinessParameter, error) {
	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.BusinessParameter{})

	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		tx = tx.Where("tenant_id = ?", orgID)
	}

	var rows []models.BusinessParameter
	if err := tx.Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list all business parameters")
	}

	out := make([]domain.BusinessParameter, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.BusinessParameter) (int64, error) {
	m := models.FromDomain(item)
	if err := r.db.Client().WithContext(ctx).Create(m).Error; err != nil {
		return 0, domainerr.Internal("failed to create business parameter")
	}
	// T1.e: dual-write de tenant_id (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE business_parameters SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, m.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set business parameter tenant")
		}
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.BusinessParameter) error {
	if err := sharedrepo.ValidateID(item.ID, "business parameter"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		existsQ := tx.Model(&models.BusinessParameter{}).Where("id = ?", item.ID)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			existsQ = existsQ.Where("tenant_id = ?", orgID)
		}
		if err := existsQ.Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("business parameter not found")
		}

		// Map ONLY the updatable fields (GORM will update Base automatically)
		updateTx := tx.Model(&models.BusinessParameter{}).
			Where("id = ?", item.ID)
		if !item.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", item.UpdatedAt)
		}
		// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			updateTx = updateTx.Where("tenant_id = ?", orgID)
		}
		result := updateTx.Updates(map[string]any{
			"key":         item.Key,
			"value":       item.Value,
			"type":        item.Type,
			"category":    item.Category,
			"description": item.Description,
		})
		if result.Error != nil {
			return domainerr.Internal("failed to update business parameter")
		}
		if result.RowsAffected == 0 {
			if !item.UpdatedAt.IsZero() {
				return domainerr.Conflict("business parameter not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", item.ID))
		}
		return nil
	})
}

func (r *Repository) ArchiveParameter(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var param models.BusinessParameter
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&param).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", id))
			}
			return domainerr.Internal("failed to get business parameter")
		}
		if param.DeletedAt.Valid {
			return domainerr.Conflict("business parameter already archived")
		}

		updates := map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": gorm.Expr("NULL"),
		}

		if err := tx.Model(&models.BusinessParameter{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return domainerr.Internal("failed to archive business parameter")
		}
		return nil
	})
}

func (r *Repository) RestoreParameter(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var param models.BusinessParameter
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&param).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", id))
			}
			return domainerr.Internal("failed to get business parameter")
		}
		if !param.DeletedAt.Valid {
			return domainerr.Conflict("business parameter is not archived")
		}

		// La reactivación puede chocar con el unique por-tenant de key (23505) → Conflict.
		if err := tx.Unscoped().Model(&models.BusinessParameter{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a business parameter with that key already exists; cannot restore")
			}
			return domainerr.Internal("failed to restore business parameter")
		}
		return nil
	})
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		existsQ := tx.Model(&models.BusinessParameter{}).Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			existsQ = existsQ.Where("tenant_id = ?", orgID)
		}
		if err := existsQ.Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("business parameter not found")
		}

		delTx := tx.Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated) — solo borra si es del tenant.
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			delTx = delTx.Where("tenant_id = ?", orgID)
		}
		result := delTx.Delete(&models.BusinessParameter{})
		if result.Error != nil {
			return domainerr.Internal("failed to delete business parameter")
		}
		if result.RowsAffected == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", id))
		}
		return nil
	})
}
