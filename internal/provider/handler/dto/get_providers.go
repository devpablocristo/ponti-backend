// Package dto define respuestas HTTP para proveedores.
package dto

import (
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

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
