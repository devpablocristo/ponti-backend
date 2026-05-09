package field

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
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
	if err := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
		Where("deleted_at IS NULL").
		Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count fields")
	}

	var list []models.Field
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Where("deleted_at IS NULL").
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
	if err := r.db.Client().WithContext(ctx).
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
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
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
		if err := tx.Unscoped().Table("fields").Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check field existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
		}

		var lotCount int64
		if err := tx.Unscoped().Model(&lotmod.Lot{}).Where("field_id = ?", id).Count(&lotCount).Error; err != nil {
			return domainerr.Internal("failed to check lots")
		}
		if lotCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("field has %d lot(s); archive or hard-delete them first", lotCount))
		}

		if err := tx.Unscoped().Delete(&models.Field{}, "id = ?", id).Error; err != nil {
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
		var f models.Field
		if err := tx.Unscoped().Where("id = ?", id).First(&f).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
			}
			return domainerr.Internal("failed to get field")
		}
		if f.DeletedAt.Valid {
			return domainerr.Conflict("field already archived")
		}

		if err := tx.Model(&models.Field{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": deletedBy,
			}).Error; err != nil {
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
		var f models.Field
		if err := tx.Unscoped().Where("id = ?", id).First(&f).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
			}
			return domainerr.Internal("failed to get field")
		}
		if !f.DeletedAt.Valid {
			return domainerr.Conflict("field is not archived")
		}

		if err := tx.Unscoped().Model(&models.Field{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return domainerr.Internal("failed to restore field")
		}
		return nil
	})
}
