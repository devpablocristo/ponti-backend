package get

import (
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

type Provider struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	ActorID *int64 `json:"actor_id,omitempty"`
}

func NewGetProvidersResponse(providersDomain []domain.Provider) []Provider {
	var providers []Provider

	for _, provider := range providersDomain {
		providers = append(
			providers,
			Provider{
				ID:      provider.ID,
				Name:    provider.Name,
				ActorID: provider.ActorID,
			},
		)
	}

	return providers
}
