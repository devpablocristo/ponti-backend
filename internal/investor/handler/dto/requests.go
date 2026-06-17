package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
)

// CreateInvestorRequest es el DTO de entrada para crear un inversor.
type CreateInvestorRequest struct {
	Name       string  `json:"name" binding:"required"`
	Percentage int     `json:"percentage"`
	TaxID      *string `json:"tax_id"` // opcional: CUIT/CUIL para el Identity Gate
}

func (r *CreateInvestorRequest) ToDomain() *domain.Investor {
	return &domain.Investor{
		Name:       r.Name,
		Percentage: r.Percentage,
		TaxID:      r.TaxID,
	}
}

// UpdateInvestorRequest es el DTO de entrada para actualizar un inversor.
type UpdateInvestorRequest struct {
	Name       string `json:"name" binding:"required"`
	Percentage int    `json:"percentage"`
}

func (r *UpdateInvestorRequest) ToDomain(id int64) *domain.Investor {
	return &domain.Investor{
		ID:         id,
		Name:       r.Name,
		Percentage: r.Percentage,
	}
}
