package dto

import (
	"time"

	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

// CreateProviderRequest es el body de POST /providers.
type CreateProviderRequest struct {
	Name string `json:"name" binding:"required,min=1"`
}

func (r CreateProviderRequest) ToDomain() *domain.Provider {
	return &domain.Provider{Name: r.Name}
}

// UpdateProviderRequest es el body de PUT /providers/:id.
type UpdateProviderRequest struct {
	Name string `json:"name" binding:"required,min=1"`
}

func (r UpdateProviderRequest) ToDomain(id int64) *domain.Provider {
	return &domain.Provider{ID: id, Name: r.Name}
}

// ProviderResponse es el DTO de salida de un proveedor individual (incl. archivado).
type ProviderResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

func ProviderFromDomain(d *domain.Provider) ProviderResponse {
	return ProviderResponse{ID: d.ID, Name: d.Name, ArchivedAt: d.ArchivedAt}
}

func NewProvidersDetailResponse(items []domain.Provider) []ProviderResponse {
	out := make([]ProviderResponse, len(items))
	for i := range items {
		out[i] = ProviderFromDomain(&items[i])
	}
	return out
}
