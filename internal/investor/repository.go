package investor

import (
	"context"
	"fmt"

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
	if err := r.db.Client().WithContext(ctx).Model(&models.Investor{}).Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count investors")
	}

	var list []models.Investor
	offset := (page - 1) * perPage
	err := r.db.Client().WithContext(ctx).
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
	if err := r.db.Client().WithContext(ctx).Unscoped().Where("id = ?", id).First(&model).Error; err != nil {
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

func (r *Repository) DeleteInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).Unscoped().Delete(&models.Investor{}, "id = ?", id)
	if result.Error != nil {
		return domainerr.Internal("failed to delete investor")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("investor with id %d does not exist", id))
	}
	return nil
}

func (r *Repository) ArchiveInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).Delete(&models.Investor{}, "id = ?", id)
	if result.Error != nil {
		return domainerr.Internal("failed to archive investor")
	}
	// Idempotente: si ya estaba archivado, RowsAffected == 0 es OK
	return nil
}

func (r *Repository) RestoreInvestor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "investor"); err != nil {
		return err
	}
	result := r.db.Client().WithContext(ctx).Unscoped().
		Model(&models.Investor{}).
		Where("id = ?", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return domainerr.Internal("failed to restore investor")
	}
	return nil
}
