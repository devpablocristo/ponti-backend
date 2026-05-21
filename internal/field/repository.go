package field

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
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

func (r *Repository) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	var fieldID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := models.FromDomain(f)
		tenantID, hasTenant, err := authz.OptionalTenantOrStrict(ctx)
		if err != nil {
			return err
		}
		if hasTenant {
			model.TenantID = tenantID
		}
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create field")
		}
		fieldID = model.ID
		for _, lot := range f.Lots {
			lotModel := lotmod.Lot{
				TenantID:       model.TenantID,
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
	base := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.Field{}), "fields").
		Where("deleted_at IS NULL")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count fields")
	}

	var list []models.Field
	offset := (page - 1) * perPage
	err := base.
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

func (r *Repository) ListArchivedFields(ctx context.Context, page, perPage int) ([]domain.Field, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Field{}).
		Where("deleted_at IS NOT NULL")
	base = authz.MaybeTenantScope(ctx, base, "fields")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived fields")
	}

	var list []models.Field
	offset := (page - 1) * perPage
	if err := base.
		Offset(offset).
		Limit(perPage).
		Order("deleted_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived fields")
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
	db0 := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "fields")
	if err := db0.
		Where("id = ? AND deleted_at IS NULL", id).
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
	updateTx := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.Field{}), "fields").
		Where("id = ?", f.ID)
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

// HardDeleteField elimina definitivamente un campo.
// Bloquea con 409 si tiene lots (activos o archivados).
func (r *Repository) HardDeleteField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		fieldDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("fields"), "fields")
		if err := fieldDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check field existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
		}
		if err := lifecycle.RequireArchived(fieldDB, "fields", "field", id); err != nil {
			return err
		}

		var lotCount int64
		lotDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&lotmod.Lot{}), "lots")
		if err := lotDB.Where("field_id = ?", id).Count(&lotCount).Error; err != nil {
			return domainerr.Internal("failed to check lots")
		}
		if lotCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("field has %d lot(s); archive or hard-delete them first", lotCount))
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "fields").Delete(&models.Field{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete field")
		}
		return nil
	})
}

// ArchiveField ejecuta un soft delete con validación.
func (r *Repository) ArchiveField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		archivedAt := time.Now()
		var f models.Field
		fieldQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "fields")
		if err := fieldQuery.Where("id = ?", id).First(&f).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
			}
			return domainerr.Internal("failed to get field")
		}
		if f.DeletedAt.Valid {
			return domainerr.Conflict("field already archived")
		}

		cause, err := lifecycle.RootCause(tx, f.TenantID, "fields", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Table("lots"), "lots").
			Where("field_id = ? AND deleted_at IS NULL", id).
			Updates(lifecycle.ArchiveUpdates(tx, "lots", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive field lots")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Field{}), "fields").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "fields", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive field")
		}
		return nil
	})
}

// RestoreField restaura un registro previamente archivado.
func (r *Repository) RestoreField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var f models.Field
		fieldQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "fields")
		if err := fieldQuery.Where("id = ?", id).First(&f).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
			}
			return domainerr.Internal("failed to get field")
		}
		if !f.DeletedAt.Valid {
			return domainerr.Conflict("field is not archived")
		}
		var projectActive int64
		if err := authz.MaybeTenantScope(ctx, tx.Table("projects"), "projects").
			Where("id = ? AND deleted_at IS NULL", f.ProjectID).
			Count(&projectActive).Error; err != nil {
			return domainerr.Internal("failed to check project")
		}
		if projectActive == 0 {
			return domainerr.Conflict("cannot restore field while project is archived")
		}
		rowState, err := lifecycle.ReadRowState(tx, "fields", id)
		if err != nil {
			return err
		}
		cause := lifecycle.CauseFromRow(rowState, "fields", id)

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Field{}), "fields").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "fields", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore field")
		}
		lotRestore := authz.MaybeTenantScope(ctx, tx.Table("lots"), "lots").
			Where("field_id = ? AND deleted_at IS NOT NULL", id)
		lotRestore = lifecycle.ApplyCauseScope(lotRestore, "lots", cause)
		if err := lotRestore.Updates(lifecycle.RestoreUpdates(tx, "lots", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore field lots")
		}
		return nil
	})
}
