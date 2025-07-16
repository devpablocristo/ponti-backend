package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
)

type Unit struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

func (d *Unit) ToDomain() *domain.Unit {
	return &domain.Unit{
		ID:   d.ID,
		Name: d.Name,
	}
}

func FromDomain(u *domain.Unit) *Unit {
	return &Unit{
		ID:   u.ID,
		Name: u.Name,
	}
}
