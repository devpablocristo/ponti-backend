package field

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
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

func (r *Repository) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	var fieldID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := sharedfilters.GuardProjectForTenant(ctx, tx, f.ProjectID); err != nil {
			return err
		}
		model := models.FromDomain(f)
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create field")
		}
		fieldID = model.ID
		for _, lot := range f.Lots {
			lotModel := lotmod.Lot{
				Name:           lot.Name,
				FieldID:        fieldID,
				Hectares:       lot.Hectares,
				PreviousCropID: lot.PreviousCrop.ID,
				CurrentCropID:  lot.CurrentCrop.ID,
				Season:         lot.Season,
			}
			if err := tx.Create(&lotModel).Error; err != nil {
				return domainerr.Internal("failed to create lot")
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return fieldID, nil
}

func (r *Repository) ListFields(ctx context.Context, page, perPage int) ([]domain.Field, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).Model(&models.Field{}).Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count fields")
	}

	var list []models.Field
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list fields")
	}

	result := make([]domain.Field, 0, len(list))
	for i := range list {
		result = append(result, *list[i].ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return nil, err
	}
	var model models.Field
	if err := r.db.Client().WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "field", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateField(ctx context.Context, f *domain.Field) error {
	if err := sharedrepo.ValidateEntity(f, "field"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(f.ID, "field"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
		Where("id = ?", f.ID)
	if cond, args := sharedfilters.TenantProjectScope(ctx); cond != "" {
		updateTx = updateTx.Where(cond, args...)
	}
	if !f.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", f.UpdatedAt)
	}
	result := updateTx.Updates(map[string]any{
		"name":          f.Name,
		"lease_type_id": f.LeaseType.ID,
	})
	if result.Error != nil {
		return domainerr.Internal("failed to update field")
	}
	if result.RowsAffected == 0 {
		if !f.UpdatedAt.IsZero() {
			return domainerr.Conflict("field not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", f.ID))
	}
	return nil
}

// DeleteField ejecuta un hard delete (permanente).
func (r *Repository) DeleteField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	delTx := r.db.Client().WithContext(ctx).
		Unscoped().
		Where("id = ?", id)
	if cond, args := sharedfilters.TenantProjectScope(ctx); cond != "" {
		delTx = delTx.Where(cond, args...)
	}
	result := delTx.Delete(&models.Field{})
	if result.Error != nil {
		return domainerr.Internal("failed to delete field")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
	}
	return nil
}

// ArchiveField ejecuta un soft delete (idempotente).
func (r *Repository) ArchiveField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	// T-child: 404 explícito si el field no es del tenant (cross-tenant). Preserva la
	// idempotencia para el propio tenant (re-archivar ya archivado -> no-op). Con flag off
	// GuardFieldForTenant es no-op.
	if err := sharedfilters.GuardFieldForTenant(ctx, r.db.Client(), id); err != nil {
		return err
	}
	archiveTx := r.db.Client().WithContext(ctx).
		Where("id = ?", id)
	if cond, args := sharedfilters.TenantProjectScope(ctx); cond != "" {
		archiveTx = archiveTx.Where(cond, args...)
	}
	result := archiveTx.Delete(&models.Field{})
	if result.Error != nil {
		return domainerr.Internal("failed to archive field")
	}
	return nil
}

// RestoreField restaura un registro previamente archivado.
func (r *Repository) RestoreField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	// T-child: 404 explícito si el field no es del tenant (cross-tenant). Preserva la
	// idempotencia para el propio tenant. Con flag off GuardFieldForTenant es no-op.
	if err := sharedfilters.GuardFieldForTenant(ctx, r.db.Client(), id); err != nil {
		return err
	}
	restoreTx := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Field{}).
		Where("id = ?", id)
	if cond, args := sharedfilters.TenantProjectScope(ctx); cond != "" {
		restoreTx = restoreTx.Where(cond, args...)
	}
	result := restoreTx.Update("deleted_at", nil)
	if result.Error != nil {
		return domainerr.Internal("failed to restore field")
	}
	return nil
}
