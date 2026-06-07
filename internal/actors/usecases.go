package actors

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
)

type RepositoryPort interface {
	Resolve(context.Context, domain.ResolveInput) (domain.ResolveResult, error)
	GetByTaxID(context.Context, string) (*domain.Actor, error)
	Search(context.Context, string, int) (domain.SearchResult, error)
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases crea los casos de uso del registro de identidad.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) Resolve(ctx context.Context, in domain.ResolveInput) (domain.ResolveResult, error) {
	return u.repo.Resolve(ctx, in)
}

func (u *UseCases) GetByTaxID(ctx context.Context, taxID string) (*domain.Actor, error) {
	return u.repo.GetByTaxID(ctx, taxID)
}

func (u *UseCases) Search(ctx context.Context, q string, limit int) (domain.SearchResult, error) {
	return u.repo.Search(ctx, q, limit)
}
