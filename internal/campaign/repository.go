package campaign

import (
	"context"

	"gorm.io/gorm"

	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	types "github.com/devpablocristo/ponti-backend/pkg/types"

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
		return 0, types.NewError(types.ErrInternal, "failed to create campaign", err)
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
			return nil, types.NewError(types.ErrInternal, "failed to list by project filter", err)
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
		if err := db.Where("id IN ?", ids).Find(&raw).Error; err != nil {
			return nil, types.NewError(types.ErrInternal, "failed to fetch filtered campaigns", err)
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
	if err := db.Find(&raw).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list campaigns", err)
	}
	out := make([]domain.Campaign, len(raw))
	for i, m := range raw {
		out[i] = *m.ToDomain()
	}
	return out, nil
}

func (r *Repository) GetCampaign(ctx context.Context, id int64) (*domain.Campaign, error) {
	var m models.Campaign
	err := r.db.Client().
		WithContext(ctx).
		First(&m, id).
		Error
	if err != nil {
		return nil, sharedrepo.HandleGormError(err, "campaign", id)
	}

	// Devolver el domain, sin exponer directamente Base
	return m.ToDomain(), nil
}
