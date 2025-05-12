package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

type Lot struct {
	ID             int64   `gorm:"primaryKey" json:"id"`
	FieldID        int64   `gorm:"not null;index" json:"field_id"`
	Name           string  `gorm:"size:100;not null" json:"name"`
	Hectares       float64 `json:"hectares"`
	PreviousCropID int64   `gorm:"not null" json:"previous_crop_id"`
	CurrentCropID  int64   `gorm:"not null" json:"current_crop_id"`
	Season         string  `gorm:"size:20;not null" json:"season"`
}

func (m *Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:             m.ID,
		FieldID:        m.FieldID,
		Name:           m.Name,
		Hectares:       m.Hectares,
		PreviousCropID: m.PreviousCropID,
		CurrentCropID:  m.CurrentCropID,
		Season:         m.Season,
	}
}

func FromDomain(d *domain.Lot) *Lot {
	return &Lot{
		FieldID:        d.FieldID,
		Name:           d.Name,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCropID,
		CurrentCropID:  d.CurrentCropID,
		Season:         d.Season,
	}
}
