package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/manager/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
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

func (r *Repository) CreateManager(ctx context.Context, m *domain.Manager) (int64, error) {
	if err := sharedrepo.ValidateEntity(m, "manager"); err != nil {
		return 0, err
	}
	model := models.FromDomain(m)
	model.Base = sharedmodels.Base{
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create manager")
	}
	return model.ID, nil
}

func (r *Repository) ListManagers(ctx context.Context, page, perPage int) ([]domain.Manager, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).
		Model(&models.Manager{}).
		Where("deleted_at IS NULL").
		Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count managers")
	}

	var list []models.Manager
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Where("deleted_at IS NULL").
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list managers")
	}

	result := make([]domain.Manager, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) ListArchivedManagers(ctx context.Context, page, perPage int) ([]domain.Manager, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Manager{}).
		Where("deleted_at IS NOT NULL")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived managers")
	}

	var list []models.Manager
	offset := (page - 1) * perPage
	if err := base.
		Offset(offset).
		Limit(perPage).
		Order("deleted_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived managers")
	}

	result := make([]domain.Manager, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return nil, err
	}
	var model models.Manager
	if err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "manager", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateManager(ctx context.Context, m *domain.Manager) error {
	if err := sharedrepo.ValidateEntity(m, "manager"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(m.ID, "manager"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Manager{}).
		Where("id = ?", m.ID)
	if !m.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", m.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomain(m))
	if result.Error != nil {
		return domainerr.Internal("failed to update manager")
	}
	if result.RowsAffected == 0 {
		if !m.UpdatedAt.IsZero() {
			return domainerr.Conflict("manager not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", m.ID))
	}
	return nil
}

// HardDeleteManager elimina definitivamente un manager.
// Bloquea con 409 si tiene asignaciones (activas o archivadas) en project_managers.
func (r *Repository) HardDeleteManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Table("managers").Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check manager existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", id))
		}

		var pmCount int64
		if err := tx.Unscoped().Table("project_managers").Where("manager_id = ?", id).Count(&pmCount).Error; err != nil {
			return domainerr.Internal("failed to check project assignments")
		}
		if pmCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("manager has %d project assignment(s); archive or remove them first", pmCount))
		}

		if err := tx.Unscoped().Delete(&models.Manager{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete manager")
		}
		return nil
	})
}

// DeleteManager queda como alias hacia HardDeleteManager para compatibilidad.
// Deprecated: usar HardDeleteManager.
func (r *Repository) DeleteManager(ctx context.Context, id int64) error {
	return r.HardDeleteManager(ctx, id)
}

// ArchiveManager archiva (soft delete) un manager con validación.
func (r *Repository) ArchiveManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m models.Manager
		if err := tx.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager %d not found", id))
			}
			return domainerr.Internal("failed to get manager")
		}
		if m.DeletedAt.Valid {
			return domainerr.Conflict("manager already archived")
		}

		if err := tx.Model(&models.Manager{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": deletedBy,
			}).Error; err != nil {
			return domainerr.Internal("failed to archive manager")
		}
		return nil
	})
}

// RestoreManager restaura un manager archivado.
func (r *Repository) RestoreManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m models.Manager
		if err := tx.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager %d not found", id))
			}
			return domainerr.Internal("failed to get manager")
		}
		if !m.DeletedAt.Valid {
			return domainerr.Conflict("manager is not archived")
		}

		if err := tx.Unscoped().Model(&models.Manager{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return domainerr.Internal("failed to restore manager")
		}
		return nil
	})
}
