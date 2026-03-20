package field

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
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
			return types.NewError(types.ErrInternal, "failed to create field", err)
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
				return types.NewError(types.ErrInternal, "failed to create lot", err)
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
		return nil, 0, types.NewError(types.ErrInternal, "failed to count fields", err)
	}

	var list []models.Field
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list fields", err)
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
	if !f.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", f.UpdatedAt)
	}
	result := updateTx.Updates(map[string]any{
		"name":          f.Name,
		"lease_type_id": f.LeaseType.ID,
	})
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update field", result.Error)
	}
	if result.RowsAffected == 0 {
		if !f.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "field not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, fmt.Sprintf("field %d not found", f.ID), nil)
	}
	return nil
}

// DeleteField ejecuta un hard delete (permanente).
func (r *Repository) DeleteField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Unscoped().
		Delete(&models.Field{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete field", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("field %d not found", id), nil)
	}
	return nil
}

// ArchiveField ejecuta un soft delete (idempotente).
func (r *Repository) ArchiveField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Field{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to archive field", result.Error)
	}
	return nil
}

// RestoreField restaura un registro previamente archivado.
func (r *Repository) RestoreField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Field{}).
		Where("id = ?", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to restore field", result.Error)
	}
	return nil
}
