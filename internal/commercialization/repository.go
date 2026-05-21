package commercialization

import (
	"context"
	"errors"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	models "github.com/devpablocristo/ponti-backend/internal/commercialization/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/commercialization/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
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

	seenProjects := make(map[int64]struct{})
	for _, item := range items {
		if _, ok := seenProjects[item.ProjectID]; ok {
			continue
		}
		if err := sharedfilters.ValidateProjectAccess(ctx, r.db.Client(), item.ProjectID); err != nil {
			return err
		}
		seenProjects[item.ProjectID] = struct{}{}
	}

	modelList := make([]models.CropCommercialization, len(items))
	tenantID, hasTenant, err := authz.OptionalTenantOrStrict(ctx)
	if err != nil {
		return err
	}
	for i, item := range items {
		modelList[i] = *models.FromDomain(&item)
		if hasTenant {
			modelList[i].TenantID = tenantID
		}
	}

	if err := r.db.Client().WithContext(ctx).Create(&modelList).Error; err != nil {
		return domainerr.Internal("failed to bulk insert crop commercializations")
	}

	return nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID int64) ([]domain.CropCommercialization, error) {
	if err := sharedfilters.ValidateProjectAccess(ctx, r.db.Client(), projectID); err != nil {
		return nil, err
	}

	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.CropCommercialization{}).
		Where("project_id = ?", projectID)
	tx = authz.MaybeTenantScope(ctx, tx, "crop_commercializations")

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
		var current models.CropCommercialization
		if err := authz.MaybeTenantScope(ctx, tx, "crop_commercializations").
			Where("id = ?", item.ID).
			First(&current).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("crop commercialization not found")
			}
			return domainerr.Internal("failed to check existence")
		}
		if err := sharedfilters.ValidateProjectAccess(ctx, tx, current.ProjectID); err != nil {
			return err
		}

		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.CropCommercialization{}), "crop_commercializations").
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
