package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

// Field is the GORM model for fields, including related lots.
type Field struct {
	ID          int64  `gorm:"primaryKey"`
	Name        string `gorm:"size:100;not null"`
	LeaseTypeID int64  `gorm:"not null;index"`
	Lots        []Lot  `gorm:"foreignKey:FieldID;constraint:OnDelete:CASCADE"`
}

// Lot is the GORM model for lots within a field, storing all attributes.
type Lot struct {
	ID int64 `gorm:"primaryKey"`
}

// ToDomain converts the Field model, including preloaded lots, into the domain Field entity.
func (m Field) ToDomain() *domain.Field {
	d := &domain.Field{
		ID:          m.ID,
		Name:        m.Name,
		LeaseTypeID: m.LeaseTypeID,
	}
	for _, lotModel := range m.Lots {
		d.Lots = append(d.Lots, lotdom.Lot{
			ID: lotModel.ID,
		})
	}
	return d
}

// FromDomainField converts a domain.Field and its lots into the GORM models for persistence.
func FromDomain(d *domain.Field) *Field {
	m := &Field{
		ID:          d.ID,
		Name:        d.Name,
		LeaseTypeID: d.LeaseTypeID,
	}
	// Map nested lots
	for _, ld := range d.Lots {
		m.Lots = append(m.Lots, Lot{
			ID: ld.ID,
		})
	}
	return m
}
