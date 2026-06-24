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

// UpdateFieldName actualiza únicamente el nombre del campo (edición desde el
// catálogo/registry unificado), sin requerir el resto del payload (lease_type, lotes).
func (r *Repository) UpdateFieldName(ctx context.Context, id int64, name string) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
		Where("id = ?", id)
	if cond, args := sharedfilters.TenantProjectScope(ctx); cond != "" {
		updateTx = updateTx.Where(cond, args...)
	}
	result := updateTx.Updates(map[string]any{"name": name})
	if result.Error != nil {
		return domainerr.Internal("failed to update field name")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("field %d not found", id))
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

// fieldScope acota el query al tenant/proyecto del caller (field es entidad hija,
// sin tenant_id propio: usa el predicado project_id de TenantProjectScope).
func fieldScope(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return func(q *gorm.DB) *gorm.DB {
		if cond, args := sharedfilters.TenantProjectScope(ctx); cond != "" {
			return q.Where(cond, args...)
		}
		return q
	}
}

// ArchiveField ejecuta un soft delete (idempotente).
func (r *Repository) ArchiveField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return sharedrepo.SoftArchive(ctx, tx, &models.Field{}, id, "field", sharedrepo.ArchiveOptions{
			Scope: fieldScope(ctx),
		})
	})
}

// RestoreField restaura un registro previamente archivado.
func (r *Repository) RestoreField(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "field"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return sharedrepo.SoftRestore(ctx, tx, &models.Field{}, id, "field", sharedrepo.ArchiveOptions{
			Scope: fieldScope(ctx),
		})
	})
}
