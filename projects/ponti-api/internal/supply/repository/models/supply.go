package models

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type Supply struct {
	ID         int64   `gorm:"primaryKey;autoIncrement;column:id"`
	ProjectID  int64   `gorm:"not null;index"`
	CampaignID int64   `gorm:"index"`
	Name       string  `gorm:"type:varchar(100);not null"`
	Unit       string  `gorm:"type:varchar(20);not null"`
	Price      float64 `gorm:"not null"`
	Category   string  `gorm:"type:varchar(50);not null"`
	Type       string  `gorm:"type:varchar(50);not null"`

	sharedmodels.Base // Audit fields
}

func (m *Supply) ToDomain() *domain.Supply {
	return &domain.Supply{
		ID:         m.ID,
		ProjectID:  m.ProjectID,
		CampaignID: m.CampaignID,
		Name:       m.Name,
		Unit:       m.Unit,
		Price:      m.Price,
		Category:   m.Category,
		Type:       m.Type,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Supply) *Supply {
	return &Supply{
		ID:         d.ID,
		ProjectID:  d.ProjectID,
		CampaignID: d.CampaignID,
		Name:       d.Name,
		Unit:       d.Unit,
		Price:      d.Price,
		Category:   d.Category,
		Type:       d.Type,
		Base: sharedmodels.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
