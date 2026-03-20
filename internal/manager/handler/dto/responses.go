package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
)

// ManagerResponse es el DTO de salida para un manager individual.
type ManagerResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// ManagerFromDomain convierte un domain.Manager a ManagerResponse.
func ManagerFromDomain(d *domain.Manager) ManagerResponse {
	return ManagerResponse{
		ID:         d.ID,
		Name:       d.Name,
		ArchivedAt: d.ArchivedAt,
	}
}

// ListManagersResponse es la respuesta paginada para el listado de managers.
type ListManagersResponse struct {
	Data     []ManagerResponse `json:"data"`
	PageInfo types.PageInfo    `json:"page_info"`
}

// NewListManagersResponse construye la respuesta paginada de managers.
func NewListManagersResponse(items []domain.Manager, page, perPage int, total int64) ListManagersResponse {
	data := make([]ManagerResponse, 0, len(items))
	for i := range items {
		data = append(data, ManagerFromDomain(&items[i]))
	}
	return ListManagersResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
