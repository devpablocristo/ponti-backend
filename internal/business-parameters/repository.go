package bparams

import (
	"context"
	"errors"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/internal/business-parameters/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/business-parameters/usecases/domain"
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

func (r *Repository) GetByKey(ctx context.Context, key string) (*domain.BusinessParameter, error) {
	var m models.BusinessParameter
	err := r.db.Client().WithContext(ctx).
		Where("key = ?", key).
		First(&m).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, types.NewError(types.ErrNotFound, "business parameter not found", nil)
	}
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get business parameter", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) ListByCategory(ctx context.Context, category string) ([]domain.BusinessParameter, error) {
	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.BusinessParameter{}).
		Where("category = ?", category)

	var rows []models.BusinessParameter
	if err := tx.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list business parameters by category", err)
	}

	out := make([]domain.BusinessParameter, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]domain.BusinessParameter, error) {
	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.BusinessParameter{})

	var rows []models.BusinessParameter
	if err := tx.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list all business parameters", err)
	}

	out := make([]domain.BusinessParameter, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.BusinessParameter) (int64, error) {
	m := models.FromDomain(item)
	if err := r.db.Client().WithContext(ctx).Create(m).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create business parameter", err)
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.BusinessParameter) error {
	if item.ID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid id", nil)
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.BusinessParameter{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "business parameter not found", nil)
		}

		// Map ONLY the updatable fields (GORM will update Base automatically)
		if err := tx.Model(&models.BusinessParameter{}).
			Where("id = ?", item.ID).
			Updates(map[string]any{
				"key":         item.Key,
				"value":       item.Value,
				"type":        item.Type,
				"category":    item.Category,
				"description": item.Description,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update business parameter", err)
		}
		return nil
	})
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return types.NewError(types.ErrInvalidID, "invalid id", nil)
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.BusinessParameter{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "business parameter not found", nil)
		}

		if err := tx.Delete(&models.BusinessParameter{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete business parameter", err)
		}
		return nil
	})
}
