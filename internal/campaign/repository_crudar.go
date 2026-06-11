package campaign

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
)

// GetArchivedCampaigns lista campañas archivadas del tenant.
func (r *Repository) GetArchivedCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	var raw []models.Campaign
	db0 := r.db.Client().WithContext(ctx).Unscoped().
		Model(&models.Campaign{}).
		Where("deleted_at IS NOT NULL")
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		db0 = db0.Where("tenant_id = ?", orgID)
	}
	if err := db0.Find(&raw).Error; err != nil {
		return nil, domainerr.Internal("failed to list archived campaigns")
	}
	out := make([]domain.Campaign, len(raw))
	for i, m := range raw {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

// UpdateCampaign renombra una campaña (dedup vía índice/trigger → 409).
func (r *Repository) UpdateCampaign(ctx context.Context, c *domain.Campaign) error {
	if err := sharedrepo.ValidateEntity(c, "campaign"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "campaign"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Campaign{}).
		Where("id = ?", c.ID)
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
	}
	result := updateTx.Updates(map[string]any{"name": c.Name, "updated_by": c.UpdatedBy})
	if result.Error != nil {
		if sharedrepo.IsUniqueViolation(result.Error) {
			return domainerr.Conflict("a campaign with that name already exists")
		}
		return domainerr.Internal("failed to update campaign")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign with id %d does not exist", c.ID))
	}
	return nil
}

// DeleteCampaign hard-borra una campaña del tenant.
func (r *Repository) DeleteCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}
	deleteTx := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id)
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		deleteTx = deleteTx.Where("tenant_id = ?", orgID)
	}
	result := deleteTx.Delete(&models.Campaign{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete campaign")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign with id %d does not exist", id))
	}
	return nil
}

// ArchiveCampaign soft-borra una campaña del tenant.
func (r *Repository) ArchiveCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m models.Campaign
		loadQ := tx.Unscoped().Where("id = ?", id)
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign %d not found", id))
			}
			return domainerr.Internal("failed to get campaign")
		}
		if m.DeletedAt.Valid {
			return domainerr.Conflict("campaign already archived")
		}
		return tx.Model(&models.Campaign{}).Where("id = ?", id).
			Updates(map[string]any{"deleted_at": time.Now(), "deleted_by": gorm.Expr("NULL")}).Error
	})
}

// RestoreCampaign reactiva una campaña archivada (dedup puede rechazar → 409).
func (r *Repository) RestoreCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m models.Campaign
		loadQ := tx.Unscoped().Where("id = ?", id)
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign %d not found", id))
			}
			return domainerr.Internal("failed to get campaign")
		}
		if !m.DeletedAt.Valid {
			return domainerr.Conflict("campaign is not archived")
		}
		if err := tx.Unscoped().Model(&models.Campaign{}).Where("id = ?", id).
			Updates(map[string]any{"deleted_at": nil, "deleted_by": nil, "updated_at": time.Now()}).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a campaign with that name already exists; cannot restore")
			}
			return domainerr.Internal("failed to restore campaign")
		}
		return nil
	})
}
