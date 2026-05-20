package crop

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
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

func (r *Repository) CreateCrop(ctx context.Context, c *domain.Crop) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "crop"); err != nil {
		return 0, err
	}
	model := models.FromDomainCrop(c)
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create crop")
	}
	return model.ID, nil
}

func (r *Repository) ListCrops(ctx context.Context, page, perPage int) ([]domain.Crop, int64, error) {
	var total int64
	base := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.Crop{}), "crops")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count crops")
	}

	var list []models.Crop
	offset := (page - 1) * perPage
	err := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "crops").
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list crops")
	}

	result := make([]domain.Crop, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetCrop(ctx context.Context, id int64) (*domain.Crop, error) {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return nil, err
	}
	var model models.Crop
	if err := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "crops").Where("id = ?", id).First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "crop", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateCrop(ctx context.Context, c *domain.Crop) error {
	if err := sharedrepo.ValidateEntity(c, "crop"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "crop"); err != nil {
		return err
	}
	updateTx := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.Crop{}), "crops").
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomainCrop(c))
	if result.Error != nil {
		return domainerr.Internal("failed to update crop")
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return domainerr.Conflict("crop not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop with id %d does not exist", c.ID))
	}
	return nil
}

func (r *Repository) ListArchivedCrops(ctx context.Context, page, perPage int) ([]domain.Crop, int64, error) {
	var total int64
	base := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Unscoped().Model(&models.Crop{}), "crops").
		Where("deleted_at IS NOT NULL")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived crops")
	}
	var list []models.Crop
	if err := base.Offset((page - 1) * perPage).Limit(perPage).Order("deleted_at DESC").Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived crops")
	}
	out := make([]domain.Crop, 0, len(list))
	for i := range list {
		out = append(out, *list[i].ToDomain())
	}
	return out, total, nil
}

func (r *Repository) ArchiveCrop(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var crop models.Crop
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "crops").Where("id = ?", id).First(&crop).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop %d not found", id))
			}
			return domainerr.Internal("failed to get crop")
		}
		if crop.DeletedAt.Valid {
			return domainerr.Conflict("crop already archived")
		}
		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, crop.TenantID, "crops", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Crop{}), "crops").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "crops", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive crop")
		}
		return nil
	})
}

func (r *Repository) RestoreCrop(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var crop models.Crop
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "crops").Where("id = ?", id).First(&crop).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop %d not found", id))
			}
			return domainerr.Internal("failed to get crop")
		}
		if !crop.DeletedAt.Valid {
			return domainerr.Conflict("crop is not archived")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Crop{}), "crops").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "crops", time.Now())).Error; err != nil {
			return domainerr.Internal("failed to restore crop")
		}
		return nil
	})
}

func (r *Repository) HardDeleteCrop(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		cropDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("crops"), "crops")
		var count int64
		if err := cropDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check crop existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop with id %d does not exist", id))
		}
		if err := lifecycle.RequireArchived(cropDB, "crops", "crop", id); err != nil {
			return err
		}
		for _, dep := range []struct {
			table string
			where string
			label string
		}{
			{"lots", "(previous_crop_id = ? OR current_crop_id = ?)", "lot"},
			{"workorders", "crop_id = ?", "work order"},
			{"crop_commercializations", "crop_id = ?", "commercialization"},
		} {
			var n int64
			query := authz.MaybeTenantScope(ctx, tx.Unscoped().Table(dep.table), dep.table)
			if dep.table == "lots" {
				query = query.Where(dep.where, id, id)
			} else {
				query = query.Where(dep.where, id)
			}
			if err := query.Count(&n).Error; err != nil {
				return domainerr.Internal("failed to check " + dep.table)
			}
			if n > 0 {
				return domainerr.Conflict(fmt.Sprintf("crop has %d %s reference(s); remove them first", n, dep.label))
			}
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "crops").Delete(&models.Crop{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete crop")
		}
		return nil
	})
}
