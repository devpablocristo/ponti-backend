package manager

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
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
	if err := r.db.Client().WithContext(ctx).Model(&models.Manager{}).Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count managers")
	}

	var list []models.Manager
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
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

func (r *Repository) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return nil, err
	}
	var model models.Manager
	if err := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id).First(&model).Error; err != nil {
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

// DeleteManager elimina físicamente (hard delete) un manager.
func (r *Repository) DeleteManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Unscoped().
		Delete(&models.Manager{}, "id = ?", id)
	if result.Error != nil {
		return domainerr.Internal("failed to delete manager")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", id))
	}
	return nil
}

// ArchiveManager archiva (soft delete) un manager. Idempotente.
func (r *Repository) ArchiveManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Manager{}, "id = ?", id)
	if result.Error != nil {
		return domainerr.Internal("failed to archive manager")
	}
	return nil
}

// RestoreManager restaura un manager archivado.
func (r *Repository) RestoreManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Manager{}).
		Where("id = ?", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return domainerr.Internal("failed to restore manager")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("manager with id %d does not exist", id))
	}
	return nil
}
