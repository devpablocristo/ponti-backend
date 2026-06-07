package crop

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
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
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		if sharedrepo.IsUniqueViolation(err) {
			return 0, domainerr.Conflict("a crop with that name already exists")
		}
		return 0, domainerr.Internal("failed to create crop")
	}
	// T1.e: dual-write de tenant_id (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE crops SET tenant_id = ? WHERE id = ? AND tenant_id IS NULL", orgID, model.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set crop tenant")
		}
	}
	return model.ID, nil
}

func (r *Repository) ListCrops(ctx context.Context, page, perPage int) ([]domain.Crop, int64, error) {
	var total int64
	countTx := r.db.Client().WithContext(ctx).Model(&models.Crop{})
	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		countTx = countTx.Where("tenant_id = ?", orgID)
	}
	if err := countTx.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count crops")
	}

	var list []models.Crop
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
	q := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — NotFound si el crop no es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	if err := q.First(&model).Error; err != nil {
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
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Crop{}).
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	// T1.e: guard de ownership (flag-gated) — solo actualiza si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		updateTx = updateTx.Where("tenant_id = ?", orgID)
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

func (r *Repository) ArchiveCrop(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var crop models.Crop
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&crop).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop %d not found", id))
			}
			return domainerr.Internal("failed to get crop")
		}
		if crop.DeletedAt.Valid {
			return domainerr.Conflict("crop already archived")
		}

		updates := map[string]any{
			"deleted_at": time.Now(),
		}
		updates["deleted_by"] = gorm.Expr("NULL")

		if err := tx.Model(&models.Crop{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
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
		loadQ := tx.Unscoped().Where("id = ?", id)
		// T1.e: guard de ownership (flag-gated).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			loadQ = loadQ.Where("tenant_id = ?", orgID)
		}
		if err := loadQ.First(&crop).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop %d not found", id))
			}
			return domainerr.Internal("failed to get crop")
		}
		if !crop.DeletedAt.Valid {
			return domainerr.Conflict("crop is not archived")
		}

		// El trigger normalize_name (dedup) dispara en la reactivación y puede
		// lanzar un unique-violation → mapear a Conflict.
		if err := tx.Unscoped().Model(&models.Crop{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			if sharedrepo.IsUniqueViolation(err) {
				return domainerr.Conflict("a crop with that name already exists; cannot restore")
			}
			return domainerr.Internal("failed to restore crop")
		}
		return nil
	})
}

func (r *Repository) DeleteCrop(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "crop"); err != nil {
		return err
	}
	deleteTx := r.db.Client().WithContext(ctx).Where("id = ?", id)
	// T1.e: guard de ownership (flag-gated) — solo borra si es del tenant.
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		deleteTx = deleteTx.Where("tenant_id = ?", orgID)
	}
	result := deleteTx.Delete(&models.Crop{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete crop")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("crop with id %d does not exist", id))
	}
	return nil
}
