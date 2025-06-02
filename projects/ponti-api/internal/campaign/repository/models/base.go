package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type Campaign struct {
	ID   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(255);not null;unique"`
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
