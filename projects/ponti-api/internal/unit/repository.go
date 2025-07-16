package unit

import (
	"context"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
	"gorm.io/gorm"
)

// GormEnginePort exposes the required DB interface.
type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListUnits(ctx context.Context) ([]domain.Unit, error) {
	var units []models.Unit
	if err := r.db.Client().WithContext(ctx).Find(&units).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list units", err)
	}
	res := make([]domain.Unit, len(units))
	for i := range units {
		res[i] = *units[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) CreateUnit(ctx context.Context, u *domain.Unit) (int64, error) {
	model := models.FromDomain(u)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create unit", err)
	}
	return model.ID, nil
}

func (r *Repository) UpdateUnit(ctx context.Context, u *domain.Unit) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Unit{}).Where("id = ?", u.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check unit existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("unit %d not found", u.ID), nil)
		}
		if err := tx.Model(&models.Unit{}).
			Where("id = ?", u.ID).
			Update("name", u.Name).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update unit", err)
		}
		return nil
	})
}

func (r *Repository) DeleteUnit(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Unit{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check unit existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("unit %d not found", id), nil)
		}
		if err := tx.Delete(&models.Unit{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete unit", err)
		}
		return nil
	})
}
