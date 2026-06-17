package dto

import (
	"time"

	"github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
)

// CreateCampaignRequest es el body de POST /campaigns.
type CreateCampaignRequest struct {
	Name string `json:"name" binding:"required,min=1"`
}

func (r CreateCampaignRequest) ToDomain() *domain.Campaign {
	return &domain.Campaign{Name: r.Name}
}

// UpdateCampaignRequest es el body de PUT /campaigns/:id.
type UpdateCampaignRequest struct {
	Name string `json:"name" binding:"required,min=1"`
}

func (r UpdateCampaignRequest) ToDomain(id int64) *domain.Campaign {
	return &domain.Campaign{ID: id, Name: r.Name}
}

// CampaignResponse es el DTO de salida (incl. archivado).
type CampaignResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	ProjectID  int64      `json:"project_id,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

func CampaignFromDomain(d *domain.Campaign) CampaignResponse {
	return CampaignResponse{ID: d.ID, Name: d.Name, ProjectID: d.ProjectID, ArchivedAt: d.ArchivedAt}
}

func NewCampaignsDetailResponse(items []domain.Campaign) []CampaignResponse {
	out := make([]CampaignResponse, len(items))
	for i := range items {
		out[i] = CampaignFromDomain(&items[i])
	}
	return out
}
