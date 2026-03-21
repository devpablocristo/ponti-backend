package bparams

import (
	"context"
	"errors"
	"fmt"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/business-parameters/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
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
		return nil, domainerr.NotFound("business parameter not found")
	}
	if err != nil {
		return nil, domainerr.Internal("failed to get business parameter")
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
		return nil, domainerr.Internal("failed to list business parameters by category")
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
		return nil, domainerr.Internal("failed to list all business parameters")
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
		return 0, domainerr.Internal("failed to create business parameter")
	}
	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.BusinessParameter) error {
	if err := sharedrepo.ValidateID(item.ID, "business parameter"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.BusinessParameter{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("business parameter not found")
		}

		// Map ONLY the updatable fields (GORM will update Base automatically)
		updateTx := tx.Model(&models.BusinessParameter{}).
			Where("id = ?", item.ID)
		if !item.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", item.UpdatedAt)
		}
		result := updateTx.Updates(map[string]any{
			"key":         item.Key,
			"value":       item.Value,
			"type":        item.Type,
			"category":    item.Category,
			"description": item.Description,
		})
		if result.Error != nil {
			return domainerr.Internal("failed to update business parameter")
		}
		if result.RowsAffected == 0 {
			if !item.UpdatedAt.IsZero() {
				return domainerr.Conflict("business parameter not found or outdated")
			}
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", item.ID))
		}
		return nil
	})
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.BusinessParameter{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("business parameter not found")
		}

		result := tx.Delete(&models.BusinessParameter{}, id)
		if result.Error != nil {
			return domainerr.Internal("failed to delete business parameter")
		}
		if result.RowsAffected == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", id))
		}
		return nil
	})
}
