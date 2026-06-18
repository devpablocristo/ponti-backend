package campaign

import (
	"context"
	"fmt"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
)

// GetArchivedCampaigns lista campañas archivadas del tenant.
func (r *Repository) GetArchivedCampaigns(ctx context.Context) ([]domain.Campaign, error) {
	var raw []models.Campaign
	db0 := r.db.Client().WithContext(ctx).Unscoped().
		Model(&models.Campaign{}).
		Where("deleted_at IS NOT NULL")
	db0 = sharedfilters.ScopeTenant(ctx, db0)
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
	updateTx = sharedfilters.ScopeTenant(ctx, updateTx)
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
	deleteTx = sharedfilters.ScopeTenant(ctx, deleteTx)
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
		return sharedrepo.SoftArchive(ctx, tx, &models.Campaign{}, id, "campaign", sharedrepo.ArchiveOptions{
			Scope: func(q *gorm.DB) *gorm.DB { return sharedfilters.ScopeTenant(ctx, q) },
		})
	})
}

// RestoreCampaign reactiva una campaña archivada (dedup puede rechazar → 409).
func (r *Repository) RestoreCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return sharedrepo.SoftRestore(ctx, tx, &models.Campaign{}, id, "campaign", sharedrepo.ArchiveOptions{
			Scope:              func(q *gorm.DB) *gorm.DB { return sharedfilters.ScopeTenant(ctx, q) },
			RestoreConflictMsg: "a campaign with that name already exists; cannot restore",
		})
	})
}
