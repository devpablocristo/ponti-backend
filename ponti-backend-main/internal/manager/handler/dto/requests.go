package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/internal/manager/usecases/domain"
)

// CreateManagerRequest es el DTO de entrada para crear un manager.
type CreateManagerRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *CreateManagerRequest) ToDomain() *domain.Manager {
	return &domain.Manager{Name: r.Name}
}

// UpdateManagerRequest es el DTO de entrada para actualizar un manager.
type UpdateManagerRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *UpdateManagerRequest) ToDomain(id int64) *domain.Manager {
	return &domain.Manager{ID: id, Name: r.Name}
}
