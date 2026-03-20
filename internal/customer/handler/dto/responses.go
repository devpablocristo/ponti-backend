package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
)

// CustomerResponse es el DTO de salida para un customer individual.
type CustomerResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// CustomerFromDomain convierte un domain.Customer a CustomerResponse.
func CustomerFromDomain(d *domain.Customer) CustomerResponse {
	return CustomerResponse{
		ID:         d.ID,
		Name:       d.Name,
		ArchivedAt: d.ArchivedAt,
	}
}

// ListedCustomerResponse es el DTO ligero para listados: sólo ID y Name.
type ListedCustomerResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ListCustomersResponse es la respuesta paginada para el listado de customers.
type ListCustomersResponse struct {
	Data     []ListedCustomerResponse `json:"data"`
	PageInfo types.PageInfo           `json:"page_info"`
}

// NewListCustomersResponse construye la respuesta paginada de customers.
func NewListCustomersResponse(items []domain.ListedCustomer, page, perPage int, total int64) ListCustomersResponse {
	data := make([]ListedCustomerResponse, 0, len(items))
	for _, c := range items {
		data = append(data, ListedCustomerResponse{
			ID:   c.ID,
			Name: c.Name,
		})
	}
	return ListCustomersResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
