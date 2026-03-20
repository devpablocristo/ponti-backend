package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/crop/usecases/domain"
)

// CreateCropRequest es el DTO de entrada para crear un crop.
type CreateCropRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *CreateCropRequest) ToDomain() *domain.Crop {
	return &domain.Crop{Name: r.Name}
}

// UpdateCropRequest es el DTO de entrada para actualizar un crop.
type UpdateCropRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r *UpdateCropRequest) ToDomain(id int64) *domain.Crop {
	return &domain.Crop{ID: id, Name: r.Name}
}
