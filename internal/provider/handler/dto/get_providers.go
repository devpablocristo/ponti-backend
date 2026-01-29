// Package dto define respuestas HTTP para proveedores.
package dto

import (
	"github.com/alphacodinggroup/ponti-backend/internal/provider/usecases/domain"
)

type Provider struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func NewGetProvidersResponse(providersDomain []domain.Provider) []Provider {
	providers := make([]Provider, len(providersDomain))
	for i, provider := range providersDomain {
		providers[i] = Provider{
			ID:   provider.ID,
			Name: provider.Name,
		}
	}

	return providers
}
