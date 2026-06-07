package campaign

import (
	"context"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"

	models "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
	projectmod "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
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

	if err := r.db.Client().
		WithContext(ctx).
		Create(model).
		Error; err != nil {
		if sharedrepo.IsUniqueViolation(err) {
			return 0, domainerr.Conflict("a campaign with that name already exists")
		}
		return 0, domainerr.Internal("failed to create campaign")
	}

	// T1.e: dual-write de tenant_id (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		if err := r.db.Client().WithContext(ctx).Exec("UPDATE campaigns SET tenant_id = ? WHERE id = ?", orgID, model.ID).Error; err != nil {
			return 0, domainerr.Internal("failed to set campaign tenant")
		}
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
		if err := db.
			Model(&projectmod.Project{}).
			Select("id, campaign_id").
			Where("customer_id = ?", customerID).
			Where("name = ?", projectName).
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
		fq := db.Where("id IN ?", ids)
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			fq = fq.Where("tenant_id = ?", orgID)
		}
		if err := fq.Find(&raw).Error; err != nil {
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
	nq := db
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		nq = nq.Where("tenant_id = ?", orgID)
	}
	if err := nq.Find(&raw).Error; err != nil {
		return nil, domainerr.Internal("failed to list campaigns")
	}
	out := make([]domain.Campaign, len(raw))
	for i, m := range raw {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	var m models.Campaign
	q := r.db.Client().WithContext(ctx)
	// T1.e: guard de ownership (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		q = q.Where("tenant_id = ?", orgID)
	}
	err := q.First(&m, id).Error
	if err != nil {
		return nil, sharedrepo.HandleGormError(err, "campaign", id)
	}

	// Devolver el domain, sin exponer directamente Base
	return m.ToDomain(), nil
}
