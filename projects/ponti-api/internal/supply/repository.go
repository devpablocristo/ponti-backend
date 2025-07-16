package supply

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
	gorm "gorm.io/gorm"
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

// --- CREATE ---
func (r *Repository) CreateSupply(ctx context.Context, s *domain.Supply) (int64, error) {
	model := models.FromDomain(s)
	// Podés setear aquí CreatedBy/UpdatedBy si viene desde el domain o contexto (opcional)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create supply", err)
	}
	return model.ID, nil
}

// --- GET ---
func (r *Repository) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	var m models.Supply
	if err := r.db.Client().WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get supply", err)
	}
	return m.ToDomain(), nil
}

// --- UPDATE ---
func (r *Repository) UpdateSupply(ctx context.Context, s *domain.Supply) error {
	model := models.FromDomain(s)
	model.ID = s.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", s.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", s.ID), nil)
		}
		updates := map[string]any{
			"name":        s.Name,
			"unit":        s.Unit,
			"price":       s.Price,
			"category":    s.Category,
			"type":        s.Type,
			"project_id":  s.ProjectID,
			"campaign_id": s.CampaignID,
			"updated_by":  s.UpdatedBy,
		}
		if err := tx.Model(&models.Supply{}).
			Where("id = ?", s.ID).
			Updates(updates).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update supply", err)
		}
		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteSupply(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), nil)
		}
		if err := tx.Delete(&models.Supply{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete supply", err)
		}
		return nil
	})
}

// --- LIST por Project ---
func (r *Repository) ListSuppliesByProject(ctx context.Context, projectID int64) ([]domain.Supply, error) {
	var supplies []models.Supply
	if err := r.db.Client().WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&supplies).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list supplies by project", err)
	}
	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}
	return res, nil
}

// --- LIST por Project y Campaign ---
func (r *Repository) ListSuppliesByProjectAndCampaign(ctx context.Context, projectID, campaignID int64) ([]domain.Supply, error) {
	var supplies []models.Supply
	if err := r.db.Client().WithContext(ctx).
		Where("project_id = ? AND campaign_id = ?", projectID, campaignID).
		Find(&supplies).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list supplies by project and campaign", err)
	}
	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}
	return res, nil
}

// --- LIST por Project o Campaign ---
func (r *Repository) ListSuppliesByProjectOrCampaign(ctx context.Context, projectID, campaignID int64) ([]domain.Supply, error) {
	var supplies []models.Supply
	if err := r.db.Client().WithContext(ctx).
		Where("project_id = ? OR campaign_id = ?", projectID, campaignID).
		Find(&supplies).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list supplies by project OR campaign", err)
	}
	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}
	return res, nil
}

// --- LIST por Campaign ---
func (r *Repository) ListSuppliesByCampaign(ctx context.Context, campaignID int64) ([]domain.Supply, error) {
	var supplies []models.Supply
	if err := r.db.Client().WithContext(ctx).
		Where("campaign_id = ?", campaignID).
		Find(&supplies).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list supplies by campaign", err)
	}
	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}
	return res, nil
}
