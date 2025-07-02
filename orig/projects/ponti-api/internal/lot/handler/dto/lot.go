package dto

import (
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type Lot struct {
	ID               int64   `json:"id,omitempty"`
	Name             string  `json:"name" binding:"required"`
	FieldID          int64   `json:"field_id" binding:"required"`
	Hectares         float64 `json:"hectares" binding:"required"`
	PreviousCropID   int64   `json:"previous_crop_id" binding:"required"`
	PreviousCropName string  `json:"previous_crop_name,omitempty"`
	CurrentCropID    int64   `json:"current_crop_id" binding:"required"`
	CurrentCropName  string  `json:"current_crop_name,omitempty"`
	Season           string  `json:"season" binding:"required"`
}

func (d *Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:           d.ID,
		Name:         d.Name,
		FieldID:      d.FieldID,
		Hectares:     d.Hectares,
		PreviousCrop: cropdom.Crop{ID: d.PreviousCropID, Name: d.PreviousCropName},
		CurrentCrop:  cropdom.Crop{ID: d.CurrentCropID, Name: d.CurrentCropName},
		Season:       d.Season,
	}
}

func FromDomain(l *domain.Lot) *Lot {
	return &Lot{
		ID:               l.ID,
		Name:             l.Name,
		FieldID:          l.FieldID,
		Hectares:         l.Hectares,
		PreviousCropID:   l.PreviousCrop.ID,
		PreviousCropName: l.PreviousCrop.Name,
		CurrentCropID:    l.CurrentCrop.ID,
		CurrentCropName:  l.CurrentCrop.Name,
		Season:           l.Season,
	}
}
