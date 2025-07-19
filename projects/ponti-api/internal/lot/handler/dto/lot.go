package dto

import (
	"time"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type Lot struct {
	ID               int64      `json:"id,omitempty"`
	Name             string     `json:"name" binding:"required"`
	FieldID          int64      `json:"field_id"`
	Hectares         float64    `json:"hectares" binding:"required"`
	PreviousCropID   int64      `json:"previous_crop_id" binding:"required"`
	PreviousCropName string     `json:"previous_crop_name,omitempty"`
	CurrentCropID    int64      `json:"current_crop_id" binding:"required"`
	CurrentCropName  string     `json:"current_crop_name,omitempty"`
	Season           string     `json:"season" binding:"required"`
	Variety          string     `json:"variety"`
	Dates            []LotDates `json:"dates"`
	Status           string     `json:"status"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (d *Lot) ToDomain() (*domain.Lot, error) {
	dates := make([]domain.LotDates, len(d.Dates))

	for i, date := range d.Dates {
		var sowingDatePtr *time.Time
		if date.SowingDate != "" {
			sowingDate, err := time.Parse("2006-01-02", date.SowingDate)
			if err != nil {
				return nil, err
			}
			sowingDatePtr = &sowingDate
		}

		var harvestDatePtr *time.Time
		if date.HarvestDate != "" {
			harvestDate, err := time.Parse("2006-01-02", date.HarvestDate)
			if err != nil {
				return nil, err
			}
			harvestDatePtr = &harvestDate
		} else {
			harvestDatePtr = nil
		}

		dates[i] = domain.LotDates{
			SowingDate:  sowingDatePtr,
			HarvestDate: harvestDatePtr,
			Sequence:    date.Sequence,
		}
	}

	return &domain.Lot{
		ID:           d.ID,
		Name:         d.Name,
		FieldID:      d.FieldID,
		Hectares:     d.Hectares,
		PreviousCrop: cropdom.Crop{ID: d.PreviousCropID, Name: d.PreviousCropName},
		CurrentCrop:  cropdom.Crop{ID: d.CurrentCropID, Name: d.CurrentCropName},
		Season:       d.Season,
		Variety:      d.Variety,
		Dates:        dates,
		Status:       d.Status,
		UpdatedAt:    d.UpdatedAt,
	}, nil
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
		Status:           l.Status,
		UpdatedAt:        l.UpdatedAt,
	}
}
