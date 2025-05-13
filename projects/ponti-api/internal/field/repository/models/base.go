package models

import (
	"time"

	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// Field is the GORM model for fields, including related lots.
type Field struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;column:id"`
	Name        string    `gorm:"size:100;not null;column:name"`
	LeaseTypeID int64     `gorm:"not null;index;column:lease_type_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;column:updated_at"`

	// One-to-many: use GORM Lot model type here.
	Lots []Lot `gorm:"foreignKey:FieldID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Lot is the GORM model for lots within a field, storing all attributes.
type Lot struct {
	ID             int64     `gorm:"primaryKey;autoIncrement;column:id"`
	FieldID        int64     `gorm:"not null;index;column:field_id"`
	Name           string    `gorm:"size:100;not null;column:name"`
	Hectares       float64   `gorm:"not null;column:hectares"`
	PreviousCropID int64     `gorm:"not null;column:previous_crop_id"`
	CurrentCropID  int64     `gorm:"not null;column:current_crop_id"`
	Season         string    `gorm:"size:20;not null;column:season"`
	CreatedAt      time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

// ToDomain converts the Field model, including preloaded lots, into the domain Field entity.
func (m Field) ToDomain() *fielddom.Field {
	d := &fielddom.Field{
		ID:          m.ID,
		Name:        m.Name,
		LeaseTypeID: m.LeaseTypeID,
	}
	for _, lotModel := range m.Lots {
		d.Lots = append(d.Lots, lotModel.ToDomain())
	}
	return d
}

// ToDomain converts the Lot GORM model into the domain Lot entity.
func (m Lot) ToDomain() lotdom.Lot {
	return lotdom.Lot{
		ID:             m.ID,
		Name:           m.Name,
		Hectares:       m.Hectares,
		PreviousCropID: m.PreviousCropID,
		CurrentCropID:  m.CurrentCropID,
		Season:         m.Season,
	}
}

// FromDomain converts a domain Field and nested lots into the GORM Field model.
func FromDomain(d *fielddom.Field) *Field {
	m := &Field{
		ID:          d.ID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseTypeID,
	}
	for _, ld := range d.Lots {
		m.Lots = append(m.Lots, Lot{
			FieldID:        d.ID,
			Name:           ld.Name,
			Hectares:       ld.Hectares,
			PreviousCropID: ld.PreviousCropID,
			CurrentCropID:  ld.CurrentCropID,
			Season:         ld.Season,
		})
	}
	return m
}
