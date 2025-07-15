package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Campaign struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
	sharedmodels.Base
}

func (c Campaign) ToDomain() *domain.Campaign {
	return &domain.Campaign{
		ID:   c.ID,
		Name: c.Name,
	}
}

func FromDomain(d *domain.Campaign) *Campaign {
	return &Campaign{
		ID:   d.ID,
		Name: d.Name,
	}
}
