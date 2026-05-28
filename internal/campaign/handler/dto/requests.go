package dto

import (
	"github.com/devpablocristo/ponti-backend/internal/campaign/usecases/domain"
)

// CreateCampaignRequest es el DTO de entrada para crear una campaign.
type CreateCampaignRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *CreateCampaignRequest) ToDomain() *domain.Campaign {
	return &domain.Campaign{Name: r.Name}
}

// UpdateCampaignRequest es el DTO de entrada para actualizar una campaign.
type UpdateCampaignRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *UpdateCampaignRequest) ToDomain(id int64) *domain.Campaign {
	return &domain.Campaign{ID: id, Name: r.Name}
}
