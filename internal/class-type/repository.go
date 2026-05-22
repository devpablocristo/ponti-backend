package classtype

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/class-type/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
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

// NOTE: ClassType ("types") is a GLOBAL catalog (no `tenant_id` column on
// the DB). Tenant scoping intentionally NOT applied — all tenants see the
// same supply types. If multi-tenant ever becomes a requirement, add the
// column via migration and reintroduce the scope here AND on the model.

func (r *Repository) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create class type")
	}
	return model.ID, nil
}

func (r *Repository) ListClassTypes(ctx context.Context, page, perPage int) ([]domain.ClassType, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).Model(&models.ClassType{}).Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count class types")
	}

	var list []models.ClassType
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list class types")
	}

	result := make([]domain.ClassType, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetClassType(ctx context.Context, id int64) (*domain.ClassType, error) {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return nil, err
	}
	var model models.ClassType
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "class type", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	if err := sharedrepo.ValidateID(c.ID, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.ClassType{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check class type existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", c.ID))
		}
		updateTx := tx.Model(&models.ClassType{}).
			Where("id = ?", c.ID)
		if !c.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
		}
		result := updateTx.Updates(map[string]any{
			"name":       c.Name,
			"updated_by": c.UpdatedBy,
		})
		if result.Error != nil {
			return domainerr.Internal("failed to update class type")
		}
		if result.RowsAffected == 0 {
			if !c.UpdatedAt.IsZero() {
				return domainerr.Conflict("class type not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", c.ID))
		}
		return nil
	})
}

func (r *Repository) ListArchivedClassTypes(ctx context.Context, page, perPage int) ([]domain.ClassType, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).Unscoped().Model(&models.ClassType{}).
		Where("deleted_at IS NOT NULL")
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived class types")
	}
	var list []models.ClassType
	if err := base.Offset((page - 1) * perPage).Limit(perPage).Order("deleted_at DESC").Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived class types")
	}
	out := make([]domain.ClassType, 0, len(list))
	for i := range list {
		out = append(out, *list[i].ToDomain())
	}
	return out, total, nil
}

func (r *Repository) ArchiveClassType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.ClassType
		if err := tx.Unscoped().Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", id))
			}
			return domainerr.Internal("failed to get class type")
		}
		if item.DeletedAt.Valid {
			return domainerr.Conflict("class type already archived")
		}
		archivedAt := time.Now()
		// Global catalog: archive batch has no tenant scope.
		cause, err := lifecycle.RootCause(tx, uuid.Nil, "types", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := tx.Model(&models.ClassType{}).
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "types", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive class type")
		}
		return nil
	})
}

func (r *Repository) RestoreClassType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.ClassType
		if err := tx.Unscoped().Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", id))
			}
			return domainerr.Internal("failed to get class type")
		}
		if !item.DeletedAt.Valid {
			return domainerr.Conflict("class type is not archived")
		}
		if err := tx.Unscoped().Model(&models.ClassType{}).
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "types", time.Now())).Error; err != nil {
			return domainerr.Internal("failed to restore class type")
		}
		return nil
	})
}

func (r *Repository) HardDeleteClassType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		typeDB := tx.Unscoped().Table("types")
		if err := typeDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check class type existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("class type %d not found", id))
		}
		if err := lifecycle.RequireArchived(typeDB, "types", "class type", id); err != nil {
			return err
		}
		for _, dep := range []struct {
			table string
			label string
		}{
			{"categories", "category"},
			{"supplies", "supply"},
		} {
			var n int64
			if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Table(dep.table), dep.table).Where("type_id = ?", id).Count(&n).Error; err != nil {
				return domainerr.Internal("failed to check " + dep.table)
			}
			if n > 0 {
				return domainerr.Conflict(fmt.Sprintf("class type has %d %s reference(s); remove them first", n, dep.label))
			}
		}
		if err := tx.Unscoped().Delete(&models.ClassType{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete class type")
		}
		return nil
	})
}
