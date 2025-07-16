package dto

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
)

type Unit struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	CreatedBy *int64 `json:"created_by,omitempty"`
	UpdatedBy *int64 `json:"updated_by,omitempty"`
}

func (d *Unit) ToDomain() *domain.Unit {
	return &domain.Unit{
		ID:   d.ID,
		Name: d.Name,
		Base: shareddomain.Base{
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

func FromDomain(u *domain.Unit) *Unit {
	return &Unit{
		ID:        u.ID,
		Name:      u.Name,
		CreatedBy: u.CreatedBy,
		UpdatedBy: u.UpdatedBy,
	}
}
