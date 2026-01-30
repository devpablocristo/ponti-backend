package manager

import (
	"context"
	"errors"
	"fmt"

	gorm "gorm.io/gorm"

	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	models "github.com/alphacodinggroup/ponti-backend/internal/manager/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/manager/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateManager(ctx context.Context, c *domain.Manager) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "manager"); err != nil {
		return 0, err
	}
	model := models.FromDomain(c)
	model.CreatedBy = c.CreatedBy
	model.UpdatedBy = c.UpdatedBy
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create manager", err)
	}
	return model.ID, nil
}

func (r *Repository) ListManagers(ctx context.Context) ([]domain.Manager, error) {
	var list []models.Manager
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list managers", err)
	}
	result := make([]domain.Manager, 0, len(list))
	for _, c := range list {
		result = append(result, *c.ToDomain())
	}
	return result, nil
}

func (r *Repository) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	var model models.Manager
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("manager with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get manager", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateManager(ctx context.Context, c *domain.Manager) error {
	if err := sharedrepo.ValidateEntity(c, "manager"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "manager"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Manager{}).
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomain(c))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update manager", result.Error)
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return types.NewError(types.ErrConflict, "manager not found or outdated", nil)
		}
		return types.NewError(types.ErrNotFound, fmt.Sprintf("manager with id %d does not exist", c.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteManager(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "manager"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Manager{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete manager", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("manager with id %d does not exist", id), nil)
	}
	return nil
}
