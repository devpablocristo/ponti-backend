// Package dto define respuestas HTTP para proveedores.
package dto

import (
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

type CreateProviderRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r CreateProviderRequest) ToDomain() *domain.Provider {
	return &domain.Provider{Name: r.Name}
}

type UpdateProviderRequest struct {
	Name string `json:"name" binding:"required"`
}

func (r UpdateProviderRequest) ToDomain(id int64) *domain.Provider {
	return &domain.Provider{ID: id, Name: r.Name}
}

type Provider struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	ActorID *int64 `json:"actor_id,omitempty"`
}

func NewGetProvidersResponse(providersDomain []domain.Provider) []Provider {
	providers := make([]Provider, len(providersDomain))
	for i, provider := range providersDomain {
		providers[i] = Provider{
			ID:      provider.ID,
			Name:    provider.Name,
			ActorID: provider.ActorID,
		}
	}

	return providers
}
