package investor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
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

func (r *Repository) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	if err := sharedrepo.ValidateEntity(inv, "investor"); err != nil {
		return 0, err
	}
	model := models.FromDomain(inv)
	model.Base = sharedmodels.Base{
		CreatedBy: inv.CreatedBy,
		UpdatedBy: inv.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create investor")
	}
	return model.ID, nil
}

func (r *Repository) ListInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	var total int64
	if err := r.db.Client().WithContext(ctx).
		Model(&models.Investor{}).
		Where("deleted_at IS NULL").
		Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count investors")
	}

	var list []models.Investor
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
		Where("deleted_at IS NULL").
		Offset(offset).
		Limit(perPage).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, 0, domainerr.Internal("failed to list investors")
	}

	result := make([]domain.Investor, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	var model models.Investor
	if err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "investor", id)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	if err := sharedrepo.ValidateEntity(inv, "investor"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(inv.ID, "investor"); err != nil {
		return err
	}
	updateTx := r.db.Client().WithContext(ctx).
		Model(&models.Investor{}).
		Where("id = ?", inv.ID)
	if !inv.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", inv.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomain(inv))
	if result.Error != nil {
		return domainerr.Internal("failed to update investor")
	}
	if result.RowsAffected == 0 {
		if !inv.UpdatedAt.IsZero() {
			return domainerr.Conflict("investor not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", inv.ID))
	}
	return nil
}

// DeleteInvestor permanece como helper interno; los handlers usan HardDeleteInvestor.
func (r *Repository) DeleteInvestor(ctx context.Context, id int64) error {
	return r.HardDeleteInvestor(ctx, id)
}

func (r *Repository) ListArchivedInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Investor{}).
		Where("deleted_at IS NOT NULL")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived investors")
	}

	var list []models.Investor
	offset := (page - 1) * perPage
	if err := base.
		Offset(offset).
		Limit(perPage).
		Order("deleted_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived investors")
	}

	result := make([]domain.Investor, 0, len(list))
	for _, m := range list {
		result = append(result, *m.ToDomain())
	}
	return result, total, nil
}

func (r *Repository) ArchiveInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv models.Investor
		if err := tx.Unscoped().Where("id = ?", id).First(&inv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor %d not found", id))
			}
			return domainerr.Internal("failed to get investor")
		}
		if inv.DeletedAt.Valid {
			return domainerr.Conflict("investor already archived")
		}

		if err := tx.Model(&models.Investor{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": deletedBy,
			}).Error; err != nil {
			return domainerr.Internal("failed to archive investor")
		}
		return nil
	})
}

func (r *Repository) RestoreInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv models.Investor
		if err := tx.Unscoped().Where("id = ?", id).First(&inv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor %d not found", id))
			}
			return domainerr.Internal("failed to get investor")
		}
		if !inv.DeletedAt.Valid {
			return domainerr.Conflict("investor is not archived")
		}

		if err := tx.Unscoped().Model(&models.Investor{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return domainerr.Internal("failed to restore investor")
		}
		return nil
	})
}

// HardDeleteInvestor elimina definitivamente un inversor.
// Bloquea con 409 si tiene registros (activos o archivados) en project_investors,
// field_investors o admin_cost_investors.
func (r *Repository) HardDeleteInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Table("investors").Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check investor existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", id))
		}

		// Validar dependientes en las tres tablas pivot (incluso archivados).
		type dep struct {
			table string
			label string
		}
		deps := []dep{
			{"project_investors", "project assignments"},
			{"field_investors", "field assignments"},
			{"admin_cost_investors", "admin cost assignments"},
		}
		for _, d := range deps {
			var n int64
			if err := tx.Unscoped().Table(d.table).Where("investor_id = ?", id).Count(&n).Error; err != nil {
				return domainerr.Internal(fmt.Sprintf("failed to check %s", d.table))
			}
			if n > 0 {
				return domainerr.Conflict(fmt.Sprintf("investor has %d %s; archive or remove them first", n, d.label))
			}
		}

		if err := tx.Unscoped().Delete(&models.Investor{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete investor")
		}
		return nil
	})
}
