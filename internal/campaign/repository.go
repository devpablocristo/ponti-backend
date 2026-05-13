package campaign

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"

	models "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	projectmod "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
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

func (r *Repository) CreateCampaign(ctx context.Context, c *domain.Campaign) (int64, error) {
	if err := sharedrepo.ValidateEntity(c, "campaign"); err != nil {
		return 0, err
	}

	// Mapear a modelo y fijar Base (CreatedBy/UpdatedBy)
	model := models.FromDomain(c)
	model.Base = sharedmodels.Base{
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}

	if err := r.db.Client().
		WithContext(ctx).
		Create(model).
		Error; err != nil {
		return 0, domainerr.Internal("failed to create campaign")
	}

	return model.ID, nil
}

func (r *Repository) ListCampaigns(ctx context.Context, customerID int64, projectName string) ([]domain.Campaign, error) {
	// Si se filtra por proyecto, obtengo campaign_id y project_id
	var filter []struct {
		ProjectID  int64 `gorm:"column:id"`
		CampaignID int64
	}

	db := r.db.Client().WithContext(ctx)

	if customerID != 0 && projectName != "" {
		projectDB := authz.MaybeTenantScope(ctx, db.Model(&projectmod.Project{}), "projects")
		if err := projectDB.
			Select("id, campaign_id").
			Where("customer_id = ?", customerID).
			Where("name = ?", projectName).
			Where("deleted_at IS NULL").
			Scan(&filter).
			Error; err != nil {
			return nil, domainerr.Internal("failed to list by project filter")
		}
	}

	// Cargo todos los campaigns (o solo los filtrados)
	var raw []models.Campaign
	if len(filter) > 0 {
		ids := make([]int64, len(filter))
		mapProject := make(map[int64]int64, len(filter))
		for i, f := range filter {
			ids[i] = f.CampaignID
			mapProject[f.CampaignID] = f.ProjectID
		}
		campaignDB := authz.MaybeTenantScope(ctx, db.Model(&models.Campaign{}), "campaigns")
		if err := campaignDB.Where("id IN ? AND deleted_at IS NULL", ids).Find(&raw).Error; err != nil {
			return nil, domainerr.Internal("failed to fetch filtered campaigns")
		}

		out := make([]domain.Campaign, 0, len(raw))
		for _, m := range raw {
			d := m.ToDomain()
			d.ProjectID = mapProject[m.ID]
			out = append(out, *d)
		}
		return out, nil
	}

	// Sin filtro
	campaignDB := authz.MaybeTenantScope(ctx, db.Model(&models.Campaign{}), "campaigns")
	if err := campaignDB.Where("deleted_at IS NULL").Find(&raw).Error; err != nil {
		return nil, domainerr.Internal("failed to list campaigns")
	}
	out := make([]domain.Campaign, len(raw))
	for i, m := range raw {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

// UpdateCampaign actualiza el nombre de una campaña existente.
func (r *Repository) UpdateCampaign(ctx context.Context, c *domain.Campaign) error {
	if err := sharedrepo.ValidateEntity(c, "campaign"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(c.ID, "campaign"); err != nil {
		return err
	}
	updateTx := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Model(&models.Campaign{}), "campaigns").
		Where("id = ?", c.ID)
	if !c.UpdatedAt.IsZero() {
		updateTx = updateTx.Where("updated_at = ?", c.UpdatedAt)
	}
	result := updateTx.Updates(models.FromDomain(c))
	if result.Error != nil {
		return domainerr.Internal("failed to update campaign")
	}
	if result.RowsAffected == 0 {
		if !c.UpdatedAt.IsZero() {
			return domainerr.Conflict("campaign not found or outdated")
		}
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign with id %d does not exist", c.ID))
	}
	return nil
}

func (r *Repository) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	var m models.Campaign
	db0 := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "campaigns")
	err := db0.
		Where("id = ? AND deleted_at IS NULL", id).
		First(&m).
		Error
	if err != nil {
		return nil, sharedrepo.HandleGormError(err, "campaign", id)
	}

	// Devolver el domain, sin exponer directamente Base
	return m.ToDomain(), nil
}

// ListArchivedCampaigns lista campañas archivadas paginadas.
func (r *Repository) ListArchivedCampaigns(ctx context.Context, page, perPage int) ([]domain.Campaign, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Campaign{}).
		Where("deleted_at IS NOT NULL")
	base = authz.MaybeTenantScope(ctx, base, "campaigns")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived campaigns")
	}

	var list []models.Campaign
	offset := (page - 1) * perPage
	if err := base.
		Offset(offset).
		Limit(perPage).
		Order("deleted_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived campaigns")
	}

	out := make([]domain.Campaign, len(list))
	for i, m := range list {
		out[i] = *m.ToDomain()
	}
	return out, total, nil
}

// ArchiveCampaign ejecuta soft delete con validación.
func (r *Repository) ArchiveCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		archivedAt := time.Now()
		var c models.Campaign
		campaignQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "campaigns")
		if err := campaignQuery.Where("id = ?", id).First(&c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign %d not found", id))
			}
			return domainerr.Internal("failed to get campaign")
		}
		if c.DeletedAt.Valid {
			return domainerr.Conflict("campaign already archived")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Campaign{}), "campaigns").
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": archivedAt,
				"deleted_by": deletedBy,
			}).Error; err != nil {
			return domainerr.Internal("failed to archive campaign")
		}
		return nil
	})
}

// RestoreCampaign restaura una campaña archivada.
func (r *Repository) RestoreCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var c models.Campaign
		campaignQuery := authz.MaybeTenantScope(ctx, tx.Unscoped(), "campaigns")
		if err := campaignQuery.Where("id = ?", id).First(&c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign %d not found", id))
			}
			return domainerr.Internal("failed to get campaign")
		}
		if !c.DeletedAt.Valid {
			return domainerr.Conflict("campaign is not archived")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Campaign{}), "campaigns").
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"deleted_by": nil,
				"updated_at": restoredAt,
			}).Error; err != nil {
			return domainerr.Internal("failed to restore campaign")
		}
		return nil
	})
}

// HardDeleteCampaign elimina definitivamente una campaña.
// Bloquea con 409 si hay proyectos (activos o archivados) referenciándola.
func (r *Repository) HardDeleteCampaign(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "campaign"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		campaignDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("campaigns"), "campaigns")
		if err := campaignDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check campaign existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("campaign %d not found", id))
		}

		var projCount int64
		projectDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&projectmod.Project{}), "projects")
		if err := projectDB.Where("campaign_id = ?", id).Count(&projCount).Error; err != nil {
			return domainerr.Internal("failed to check projects")
		}
		if projCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("campaign has %d project(s); archive or hard-delete them first", projCount))
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "campaigns").Delete(&models.Campaign{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete campaign")
		}
		return nil
	})
}
