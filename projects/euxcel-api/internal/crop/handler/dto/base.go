package dto

import (
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/crop/usecases/domain"
)

// Crop is the DTO for a specific crop.
type Crop struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Season string `json:"season"`
	LotID  int64  `json:"lot_id"`
}

// ToDomain converts the DTO Crop to the domain entity.
func (c Crop) ToDomain() *domain.Crop {
	return &domain.Crop{
		ID:     c.ID,
		Name:   c.Name,
		Season: c.Season,
		LotID:  c.LotID,
	}
}

// FromDomain converts a domain Crop to the DTO.
func FromDomain(d domain.Crop) *Crop {
	return &Crop{
		ID:     d.ID,
		Name:   d.Name,
		Season: d.Season,
		LotID:  d.LotID,
	}
}
