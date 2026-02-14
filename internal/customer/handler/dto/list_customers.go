package dto

import (
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domain "github.com/alphacodinggroup/ponti-backend/internal/customer/usecases/domain"
)

// ListedCustomer es el DTO ligero para listados: sólo ID y Name.
type ListedCustomer struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ListCustomersResponse es la respuesta paginada.
type ListCustomersResponse struct {
	Items    []ListedCustomer `json:"items"`
	PageInfo types.PageInfo   `json:"page_info"`
}

// NewListCustomersResponse construye la respuesta paginada.
func NewListCustomersResponse(
	items []domain.ListedCustomer,
	page, perPage int,
	total int64,
) ListCustomersResponse {
	out := make([]ListedCustomer, len(items))
	for i, c := range items {
		out[i] = ListedCustomer{
			ID:   c.ID,
			Name: c.Name,
		}
	}

	return ListCustomersResponse{
		Items:    out,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
