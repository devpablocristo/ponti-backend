package dto

import (
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

// Estructura simple de supply para listados rápidos
type ListedSupply struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Respuesta principal de listado de supplies
type ListSuppliesResponse struct {
	Data     []ListedSupply `json:"data"`
	PageInfo types.PageInfo `json:"page_info"`
}

// Constructor para la respuesta paginada de supplies
func NewListSuppliesResponse(
	items []domain.Supply,
	page, perPage int,
	total int64,
) ListSuppliesResponse {
	out := make([]ListedSupply, len(items))
	for i, s := range items {
		out[i] = ListedSupply{
			ID:   s.ID,
			Name: s.Name,
		}
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))

	return ListSuppliesResponse{
		Data: out,
		PageInfo: types.PageInfo{
			PerPage: perPage,
			Page:    page,
			MaxPage: maxPage,
			Total:   total,
		},
	}
}
