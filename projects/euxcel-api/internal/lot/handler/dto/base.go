package dto

import (
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// Lot is the DTO for a specific lot.
type Lot struct {
	ID      int64   `json:"lot_id"`
	FieldID int64   `json:"field_id"`
	LotName string  `json:"lot_name"`
	Area    float64 `json:"area"`
}

// ToDomain converts the DTO Lot to the domain entity.
func (l Lot) ToDomain() *domain.Lot {
	return &domain.Lot{
		ID:      l.ID,
		FieldID: l.FieldID,
		LotName: l.LotName,
		Area:    l.Area,
	}
}

// FromDomain converts a domain Lot to the DTO.
func FromDomain(d domain.Lot) *Lot {
	return &Lot{
		ID:      d.ID,
		FieldID: d.FieldID,
		LotName: d.LotName,
		Area:    d.Area,
	}
}
