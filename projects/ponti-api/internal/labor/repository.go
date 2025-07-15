package labor

import (
	"context"
	"fmt"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
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

func (r *Repository) CreateLabor(ctx context.Context, inv *domain.Labor) (int64, error) {
	if inv == nil {
		return 0, types.NewError(types.ErrValidation, "labor is nil", nil)
	}
	model := models.FromDomain(inv)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create labor", err)
	}
	return model.ID, nil
}

func (r *Repository) deleteLabor(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Labor{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete investor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("labor with id %d does not exist", id), nil)
	}
	return nil
}

func (r *Repository) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	if labor == nil {
		return types.NewError(types.ErrValidation, "investor is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Where("id = ?", labor.ID).
		Updates(models.FromDomain(labor))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update labor", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("investor with id %d does not exist", labor.ID), nil)
	}
	return nil
}

func (r *Repository) ListLabor(ctx context.Context, page, perPage int) ([]domain.ListedLabor, int64, error) {
	var list []models.Labor
	var total int64

	db0 := r.db.Client().WithContext(ctx).Model(&models.Labor{})

	// Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count labors", err)
	}

	if err := db0.
		Select("id, name, contractor_name, price, category_id").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list labor", err)
	}

	// Mapear a dominio ligero
	labors := make([]domain.ListedLabor, len(list))
	for i, labor := range list {
		labors[i] = domain.ListedLabor{
			ID:              labor.ID,
			Name:            labor.Name,
			Price:           labor.Price,
			ContractorName:  labor.ContractorName,
			LaborCategoryId: labor.LaborCategoryID,
		}
	}

	return labors, total, nil
}
