package models

import (
	cropdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// Lot is the GORM model for a land parcel, storing only foreign-key references.
type Lot struct {
	ID             int64   `gorm:"primaryKey"`
	Name           string  `gorm:"size:100;not null"`
	Hectares       float64 `gorm:"not null"`
	PreviousCropID int64   `gorm:"not null;index"`
	CurrentCropID  int64   `gorm:"not null;index"`
	Season         string  `gorm:"size:20;not null"`
}

// ToDomain maps the GORM model to the domain.Lot, using cropdom.Crop for crop references.
func (m *Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:           m.ID,
		Name:         m.Name,
		Hectares:     m.Hectares,
		PreviousCrop: cropdom.Crop{ID: m.PreviousCropID},
		CurrentCrop:  cropdom.Crop{ID: m.CurrentCropID},
		Season:       m.Season,
	}
}

// FromDomain maps the domain.Lot to the GORM model, extracting crop IDs.
func FromDomain(d *domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCrop.ID,
		CurrentCropID:  d.CurrentCrop.ID,
		Season:         d.Season,
	}
}
