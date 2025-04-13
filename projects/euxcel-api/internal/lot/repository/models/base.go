package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// Lot represents the GORM model for a lot.
type Lot struct {
	ID      int64   `gorm:"primaryKey"`
	FieldID int64   `gorm:"not null"`
	LotName string  `gorm:"type:varchar(100);not null"`
	Area    float64 `gorm:"not null"`
}

// ToDomain converts the Lot model to the domain entity.
func (l Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:      l.ID,
		FieldID: l.FieldID,
		LotName: l.LotName,
		Area:    l.Area,
	}
}

// FromDomainLot converts a domain Lot to the GORM model.
func FromDomainLot(d *domain.Lot) *Lot {
	return &Lot{
		ID:      d.ID,
		FieldID: d.FieldID,
		LotName: d.LotName,
		Area:    d.Area,
	}
}
