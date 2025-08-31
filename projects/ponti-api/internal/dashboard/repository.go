package dashboard

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
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

func (r *Repository) CreateDashboard(ctx context.Context, d *domain.Dashboard) (int64, error) {
	if d == nil {
		return 0, types.NewError(types.ErrValidation, "dashboard is nil", nil)
	}
	model := models.FromDomainDashboard(d)
	model.Base = sharedmodels.Base{
		CreatedBy: d.CreatedBy,
		UpdatedBy: d.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create dashboard", err)
	}
	return model.ID, nil
}

func (r *Repository) ListDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	var list []models.Dashboard
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list dashboards", err)
	}
	result := make([]domain.Dashboard, 0, len(list))
	for _, d := range list {
		result = append(result, *d.ToDomain())
	}
	return result, nil
}

func (r *Repository) GetDashboard(ctx context.Context, id int64) (*domain.Dashboard, error) {
	var model models.Dashboard
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("dashboard with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateDashboard(ctx context.Context, d *domain.Dashboard) error {
	if d == nil {
		return types.NewError(types.ErrValidation, "dashboard is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Dashboard{}).
		Where("id = ?", d.ID).
		Updates(models.FromDomainDashboard(d))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update dashboard", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("dashboard with id %d does not exist", d.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteDashboard(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Dashboard{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete dashboard", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("dashboard with id %d does not exist", id), nil)
	}
	return nil
}
