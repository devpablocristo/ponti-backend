package bparams

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/devpablocristo/platform/persistence/gorm/go/tenancy"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
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
	err := tenancy.Scope(ctx, r.db.Client().WithContext(ctx), "business_parameters").
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
	tx = tenancy.Scope(ctx, tx, "business_parameters")

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
	tx = tenancy.Scope(ctx, tx, "business_parameters")

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

func (r *Repository) ListArchived(ctx context.Context) ([]domain.BusinessParameter, error) {
	tx := r.db.Client().
		WithContext(ctx).
		Unscoped().
		Model(&models.BusinessParameter{}).
		Where("deleted_at IS NOT NULL")
	tx = tenancy.Scope(ctx, tx, "business_parameters")

	var rows []models.BusinessParameter
	if err := tx.Order("deleted_at DESC").Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list archived business parameters")
	}

	out := make([]domain.BusinessParameter, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) Create(ctx context.Context, item *domain.BusinessParameter) (int64, error) {
	m := models.FromDomain(item)
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		m.TenantID = tenantID
	}
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
		if err := tenancy.Scope(ctx, tx.Model(&models.BusinessParameter{}), "business_parameters").Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("business parameter not found")
		}

		// Map ONLY the updatable fields (GORM will update Base automatically)
		updateTx := tenancy.Scope(ctx, tx.Model(&models.BusinessParameter{}), "business_parameters").
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

func (r *Repository) Archive(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.BusinessParameter
		if err := tenancy.Scope(ctx, tx.Unscoped(), "business_parameters").Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("business parameter not found")
			}
			return domainerr.Internal("failed to get business parameter")
		}
		if item.DeletedAt.Valid {
			return domainerr.Conflict("business parameter already archived")
		}
		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, item.TenantID, "business_parameters", id, nil, deletedBy)
		if err != nil {
			return err
		}
		result := tenancy.Scope(ctx, tx.Model(&models.BusinessParameter{}), "business_parameters").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "business_parameters", archivedAt, deletedBy, cause))
		if result.Error != nil {
			return domainerr.Internal("failed to archive business parameter")
		}
		if result.RowsAffected == 0 {
			return domainerr.NotFound("business parameter not found")
		}
		return nil
	})
}

func (r *Repository) Restore(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.BusinessParameter
		if err := tenancy.Scope(ctx, tx.Unscoped(), "business_parameters").Where("id = ?", id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("business parameter not found")
			}
			return domainerr.Internal("failed to get business parameter")
		}
		if !item.DeletedAt.Valid {
			return domainerr.Conflict("business parameter is not archived")
		}
		result := tenancy.Scope(ctx, tx.Unscoped().Model(&models.BusinessParameter{}), "business_parameters").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "business_parameters", time.Now()))
		if result.Error != nil {
			return domainerr.Internal("failed to restore business parameter")
		}
		if result.RowsAffected == 0 {
			return domainerr.NotFound("business parameter not found")
		}
		return nil
	})
}

func (r *Repository) HardDelete(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "business parameter"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		base := tenancy.Scope(ctx, tx.Unscoped().Table("business_parameters"), "business_parameters")
		var count int64
		if err := base.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("business parameter not found")
		}
		if err := lifecycle.RequireArchived(base, "business_parameters", "business parameter", id); err != nil {
			return err
		}
		result := tenancy.Scope(ctx, tx.Unscoped(), "business_parameters").Delete(&models.BusinessParameter{}, id)
		if result.Error != nil {
			return domainerr.Internal("failed to hard delete business parameter")
		}
		if result.RowsAffected == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("business parameter %d not found", id))
		}
		return nil
	})
}
