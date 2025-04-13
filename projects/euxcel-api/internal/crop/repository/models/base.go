package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop/usecases/domain"
)

// Crop represents the GORM model for a crop.
type Crop struct {
	ID     int64  `gorm:"primaryKey"`
	Name   string `gorm:"type:varchar(100);not null"`
	Season string `gorm:"type:varchar(50);not null"`
	LotID  int64  `gorm:"not null"`
}

// ToDomain converts the Crop model to the domain entity.
func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:     c.ID,
		Name:   c.Name,
		Season: c.Season,
		LotID:  c.LotID,
	}
}

// FromDomainCrop converts a domain Crop entity to the GORM model.
func FromDomainCrop(d *domain.Crop) *Crop {
	return &Crop{
		ID:     d.ID,
		Name:   d.Name,
		Season: d.Season,
		LotID:  d.LotID,
	}
}
