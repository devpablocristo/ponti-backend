package models

import (
	cropmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	sheredmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
)

type Lot struct {
	ID             int64   `gorm:"primaryKey;autoIncrement;column:id"`
	Name           string  `gorm:"type:varchar(100);not null;column:name"`
	FieldID        int64   `gorm:"not null;index;constraint:OnDelete:CASCADE;"`
	Hectares       float64 `gorm:"not null;column:hectares"`
	PreviousCropID int64   `gorm:"not null;index;column:previous_crop_id"`
	CurrentCropID  int64   `gorm:"not null;index;column:current_crop_id"`
	Season         string  `gorm:"size:20;not null;column:season"`
	sheredmodels.Base
	Variety      string       `gorm:"size:20;not null;column:variety"`
	PreviousCrop cropmod.Crop `gorm:"foreignKey:PreviousCropID;references:ID"`
	CurrentCrop  cropmod.Crop `gorm:"foreignKey:CurrentCropID;references:ID"`
}

func (m *Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:       m.ID,
		Name:     m.Name,
		FieldID:  m.FieldID,
		Hectares: m.Hectares,
		PreviousCrop: cropdom.Crop{
			ID: m.PreviousCropID,
		},
		CurrentCrop: cropdom.Crop{
			ID: m.CurrentCropID,
		},
		Season:  m.Season,
		Variety: m.Variety,
	}
}

func FromDomain(d *domain.Lot) *Lot {
	return &Lot{
		ID:             d.ID,
		Name:           d.Name,
		FieldID:        d.FieldID,
		Hectares:       d.Hectares,
		PreviousCropID: d.PreviousCrop.ID,
		CurrentCropID:  d.CurrentCrop.ID,
		Season:         d.Season,
	}
}
