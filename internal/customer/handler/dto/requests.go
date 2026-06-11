package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
)

// CreateCustomerRequest es el DTO de entrada para crear un customer.
type CreateCustomerRequest struct {
	Name  string  `json:"name" binding:"required"`
	TaxID *string `json:"tax_id"` // opcional: CUIT/CUIL para el Identity Gate
}

func (r *CreateCustomerRequest) ToDomain() *domain.Customer {
	return &domain.Customer{Name: r.Name, TaxID: r.TaxID}
}

// UpdateCustomerRequest es el DTO de entrada para actualizar un customer.
type UpdateCustomerRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *UpdateCustomerRequest) ToDomain(id int64) *domain.Customer {
	return &domain.Customer{ID: id, Name: r.Name}
}
