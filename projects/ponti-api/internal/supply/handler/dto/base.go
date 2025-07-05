package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type Supply struct {
	ID         int64   `json:"id,omitempty"`
	ProjectID  int64   `json:"project_id"`
	CampaignID int64   `json:"campaign_id,omitempty"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	Price      float64 `json:"price"`
	Category   string  `json:"category"`
	Type       string  `json:"type"`
}

func (d *Supply) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		CampaignID: d.CampaignID,
		Name:       d.Name,
		Unit:       d.Unit,
		Price:      d.Price,
		Category:   d.Category,
		Type:       d.Type,
	}
}

func FromDomain(s *domain.Supply) *Supply {
	return &Supply{
		ID:         s.ID,
		ProjectID:  s.ProjectID,
		CampaignID: s.CampaignID,
		Name:       s.Name,
		Unit:       s.Unit,
		Price:      s.Price,
		Category:   s.Category,
		Type:       s.Type,
	}
}
