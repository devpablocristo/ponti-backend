package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/category/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
)

// CategoryResponse es el DTO de salida para una categoría individual.
type CategoryResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TypeID int64  `json:"type_id"`
}

// CategoryFromDomain convierte un domain.Category a CategoryResponse.
func CategoryFromDomain(d *domain.Category) CategoryResponse {
	return CategoryResponse{ID: d.ID, Name: d.Name, TypeID: d.TypeID}
}

// ListCategoriesResponse es la respuesta paginada para el listado de categorías.
type ListCategoriesResponse struct {
	Data     []CategoryResponse `json:"data"`
	PageInfo types.PageInfo     `json:"page_info"`
}

// NewListCategoriesResponse construye la respuesta paginada de categorías.
func NewListCategoriesResponse(items []domain.Category, page, perPage int, total int64) ListCategoriesResponse {
	data := make([]CategoryResponse, 0, len(items))
	for i := range items {
		data = append(data, CategoryFromDomain(&items[i]))
	}
	return ListCategoriesResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
