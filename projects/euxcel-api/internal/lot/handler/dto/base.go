package dto

import (
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// Lot matches the POST/PUT payload and includes FieldID.
type Lot struct {
	FieldID        int64   `json:"field_id"`
	Name           string  `json:"name"`
	Hectares       float64 `json:"hectares"`
	PreviousCropID int64   `json:"previous_crop_id"`
	CurrentCropID  int64   `json:"current_crop_id"`
	Season         string  `json:"season"`
}

// ToDomain converts the DTO into a domain.Lot.
func (p Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		FieldID:        p.FieldID,
		Name:           p.Name,
		Hectares:       p.Hectares,
		PreviousCropID: p.PreviousCropID,
		CurrentCropID:  p.CurrentCropID,
		Season:         p.Season,
	}
}

// FromDomain converts a domain.Lot into a DTO.
func FromDomain(d domain.Lot) *Lot {
	return &Lot{
		FieldID:        d.FieldID,
		Name:           d.Name,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCropID,
		CurrentCropID:  d.CurrentCropID,
		Season:         d.Season,
	}
}
