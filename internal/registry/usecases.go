package registry

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/registry/usecases/domain"
)

type RepositoryPort interface {
	SearchRegistry(ctx context.Context, q, typ, status string, page, perPage int) (domain.RegistryResult, error)
	SetAliases(ctx context.Context, actorID int64, aliases []string) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) SearchRegistry(ctx context.Context, q, typ, status string, page, perPage int) (domain.RegistryResult, error) {
	return u.repo.SearchRegistry(ctx, q, typ, status, page, perPage)
}

func (u *UseCases) SetAliases(ctx context.Context, actorID int64, aliases []string) error {
	return u.repo.SetAliases(ctx, actorID, aliases)
}
