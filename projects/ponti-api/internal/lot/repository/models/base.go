package models

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// Lot is the GORM model for a land parcel, storing only foreign-key references.
type Lot struct {
	ID             int64     `gorm:"primaryKey"`
	Name           string    `gorm:"size:100;not null"`
	FieldID        int64     `gorm:"not null;index;column:field_id"`
	Hectares       float64   `gorm:"not null"`
	PreviousCropID int64     `gorm:"not null;index"`
	CurrentCropID  int64     `gorm:"not null;index"`
	Season         string    `gorm:"size:20;not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

// ToDomain maps the GORM model to the domain.Lot, using cropdom.Crop for crop references.
func (m *Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:             m.ID,
		Name:           m.Name,
		FieldID:        m.FieldID,
		Hectares:       m.Hectares,
		PreviousCropID: m.PreviousCropID,
		CurrentCropID:  m.CurrentCropID,
		Season:         m.Season,
	}
}

// FromDomain maps the domain.Lot to the GORM model, extracting crop IDs.
func FromDomain(d *domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		FieldID:        d.FieldID,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCropID,
		CurrentCropID:  d.CurrentCropID,
		Season:         d.Season,
	}
}
