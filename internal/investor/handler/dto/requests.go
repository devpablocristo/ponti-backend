package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/internal/investor/usecases/domain"
)

// CreateInvestorRequest es el DTO de entrada para crear un inversor.
type CreateInvestorRequest struct {
	Name       string `json:"name" binding:"required"`
	Percentage int    `json:"percentage"`
}

func (r *CreateInvestorRequest) ToDomain() *domain.Investor {
	return &domain.Investor{
		Name:       r.Name,
		Percentage: r.Percentage,
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
