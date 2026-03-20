package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
)

// CreateCategoryRequest es el DTO de entrada para crear una categoría.
type CreateCategoryRequest struct {
	Name   string `json:"name" binding:"required"`
	TypeID int64  `json:"type_id" binding:"required"`
}

func (r *CreateCategoryRequest) ToDomain() *domain.Category {
	return &domain.Category{Name: r.Name, TypeID: r.TypeID}
}

// UpdateCategoryRequest es el DTO de entrada para actualizar una categoría.
type UpdateCategoryRequest struct {
	Name   string `json:"name" binding:"required"`
	TypeID int64  `json:"type_id" binding:"required"`
}

func (r *UpdateCategoryRequest) ToDomain(id int64) *domain.Category {
	return &domain.Category{ID: id, Name: r.Name, TypeID: r.TypeID}
}
