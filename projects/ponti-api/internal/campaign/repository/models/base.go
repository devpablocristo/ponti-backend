package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
)

type Campaign struct {
	ID   int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Name string `gorm:"type:varchar(100);not null"`
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
