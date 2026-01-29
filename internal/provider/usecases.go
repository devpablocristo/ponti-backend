package provider

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/internal/provider/usecases/domain"
)

type RepositoryPort interface {
	GetProviders(context.Context) ([]domain.Provider, error)
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases crea una instancia de los casos de uso para Provider.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetProviders(ctx context.Context) ([]domain.Provider, error) {
	return u.repo.GetProviders(ctx)
}
