package commercialization

import (
	"context"
	"time"

	"github.com/devpablocristo/saas-core/shared/domainerr"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/commercialization/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/commercialization/usecases/domain"
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

func (r *Repository) CreateBulk(ctx context.Context, items []domain.CropCommercialization) error {
	if len(items) == 0 {
		return nil
	}

	modelList := make([]models.CropCommercialization, len(items))
	for i, item := range items {
		modelList[i] = *models.FromDomain(&item)
	}

	if err := r.db.Client().WithContext(ctx).Create(&modelList).Error; err != nil {
		return domainerr.Internal("failed to bulk insert crop commercializations")
	}

	return nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID int64) ([]domain.CropCommercialization, error) {

	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.CropCommercialization{}).
		Where("project_id = ?", projectID)

	var rows []models.CropCommercialization

	if err := tx.Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list crop commercialization")
	}

	if len(rows) == 0 {
		return nil, domainerr.NotFound("no commercializations found for this project")
	}

	out := make([]domain.CropCommercialization, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}

	return out, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.CropCommercialization) error {
	if item.ID == 0 {
		return domainerr.Validation("invalid ID")
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.CropCommercialization{}).
			Where("id = ?", item.ID).
			Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check existence")
		}
		if count == 0 {
			return domainerr.NotFound("crop commercialization not found")
		}

		if err := tx.Model(&models.CropCommercialization{}).
			Where("id = ?", item.ID).
			Updates(map[string]any{
				"crop_id":         item.CropID,
				"board_price":     item.BoardPrice,
				"freight_cost":    item.FreightCost,
				"commercial_cost": item.CommercialCost,
				"net_price":       item.NetPrice,
				"updated_at":      time.Now(),
				"updated_by":      item.UpdatedBy,
			}).Error; err != nil {
			return domainerr.Internal("failed to update crop commercialization")
		}
		return nil
	})
}
