package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
)

// CreateCustomerRequest es el DTO de entrada para crear un customer.
type CreateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	ActorID *int64 `json:"actor_id,omitempty"`
}

func (r *CreateCustomerRequest) ToDomain() *domain.Customer {
	return &domain.Customer{Name: r.Name, ActorID: r.ActorID}
}

// UpdateCustomerRequest es el DTO de entrada para actualizar un customer.
type UpdateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	ActorID *int64 `json:"actor_id,omitempty"`
}

func (r *UpdateCustomerRequest) ToDomain(id int64) *domain.Customer {
	return &domain.Customer{ID: id, Name: r.Name, ActorID: r.ActorID}
}
