package models

import (
	"time"

	cropmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	fieldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// Lot is the GORM model for a land parcel, storing only foreign-key references.
type Lot struct {
	ID             int64          `gorm:"primaryKey;autoIncrement"`
	Name           string         `gorm:"size:100;not null"`
	FieldID        int64          `gorm:"not null;index;column:field_id"`
	Field          fieldmod.Field `gorm:"foreignKey:FieldID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Hectares       float64        `gorm:"not null"`
	PreviousCropID int64          `gorm:"not null;index"`
	PreviousCrop   cropmod.Crop   `gorm:"foreignKey:PreviousCropID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CurrentCropID  int64          `gorm:"not null;index"`
	CurrentCrop    cropmod.Crop   `gorm:"foreignKey:CurrentCropID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Season         string         `gorm:"size:20;not null"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime;column:updated_at"`
}

func (m *Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:           m.ID,
		Name:         m.Name,
		FieldID:      m.FieldID,
		Hectares:     m.Hectares,
		PreviousCrop: cropdom.Crop{ID: m.PreviousCropID},
		CurrentCrop:  cropdom.Crop{ID: m.CurrentCropID},
		Season:       m.Season,
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
