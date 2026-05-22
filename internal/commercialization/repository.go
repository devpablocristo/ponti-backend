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
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
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

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range items {
			if err := assertCommercializationReferencesActive(tx, &items[i]); err != nil {
				return err
			}
		}
		if err := tx.Create(&modelList).Error; err != nil {
			return domainerr.Internal("failed to bulk insert crop commercializations")
		}
		return nil
	})
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

	// Empty list is a valid state: a project that hasn't loaded its
	// commercialization yet shows the form to create it. Returning 404 here
	// (the previous behavior) made the FE log a noisy error on every visit
	// to `Cargar Comercialización` for fresh projects.
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
		// Use current.ProjectID as source of truth — payload only changes crop
		// and pricing fields. CropID may differ, validate the new value.
		toValidate := domain.CropCommercialization{
			ProjectID: current.ProjectID,
			CropID:    item.CropID,
		}
		if err := assertCommercializationReferencesActive(tx, &toValidate); err != nil {
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

// assertCommercializationReferencesActive blocks Create/Update of a
// commercialization that references an archived project or crop. Both are
// required parents; if either is archived the commercialization row would
// violate the hierarchical invariant.
func assertCommercializationReferencesActive(tx *gorm.DB, c *domain.CropCommercialization) error {
	if c == nil {
		return nil
	}
	refs := []lifecycle.ActiveRef{
		{Table: "projects", Label: "project", ID: c.ProjectID},
		{Table: "crops", Label: "crop", ID: c.CropID},
	}
	return lifecycle.RequireAllActive(tx, refs)
}
