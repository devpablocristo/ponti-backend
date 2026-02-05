package classtype

import (
	"context"
	"fmt"

	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/internal/class-type/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/class-type/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	"gorm.io/gorm"
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

func (r *Repository) ListClassTypes(ctx context.Context) ([]domain.ClassType, error) {
	var classTypes []models.ClassType
	if err := r.db.Client().WithContext(ctx).Find(&classTypes).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list class types", err)
	}
	res := make([]domain.ClassType, len(classTypes))
	for i := range classTypes {
		res[i] = *classTypes[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create class type", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	if err := sharedrepo.ValidateID(c.ID, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.ClassType{}).Where("id = ?", c.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check class type existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("class type %d not found", c.ID), nil)
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
			return types.NewError(types.ErrInternal, "failed to update class type", result.Error)
		}
		if result.RowsAffected == 0 {
			if !c.UpdatedAt.IsZero() {
				return types.NewError(types.ErrConflict, "class type not found or outdated", nil)
			}
			return types.NewError(types.ErrNotFound, fmt.Sprintf("class type %d not found", c.ID), nil)
		}
		return nil
	})
}

func (r *Repository) DeleteClassType(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "class type"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.ClassType{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check class type existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("class type %d not found", id), nil)
		}
		result := tx.Delete(&models.ClassType{}, id)
		if result.Error != nil {
			return types.NewError(types.ErrInternal, "failed to delete class type", result.Error)
		}
		if result.RowsAffected == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("class type %d not found", id), nil)
		}
		return nil
	})
}
