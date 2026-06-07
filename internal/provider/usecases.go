package provider

import (
	"context"

	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

type RepositoryPort interface {
	GetProviders(context.Context) ([]domain.Provider, error)
	GetArchivedProviders(context.Context) ([]domain.Provider, error)
	CreateProvider(context.Context, *domain.Provider) (int64, error)
	GetProvider(context.Context, int64) (*domain.Provider, error)
	UpdateProvider(context.Context, *domain.Provider) error
	DeleteProvider(context.Context, int64) error
	ArchiveProvider(context.Context, int64) error
	RestoreProvider(context.Context, int64) error
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

func (u *UseCases) GetArchivedProviders(ctx context.Context) ([]domain.Provider, error) {
	return u.repo.GetArchivedProviders(ctx)
}

func (u *UseCases) CreateProvider(ctx context.Context, p *domain.Provider) (int64, error) {
	return u.repo.CreateProvider(ctx, p)
}

func (u *UseCases) GetProvider(ctx context.Context, id int64) (*domain.Provider, error) {
	return u.repo.GetProvider(ctx, id)
}

func (u *UseCases) UpdateProvider(ctx context.Context, p *domain.Provider) error {
	return u.repo.UpdateProvider(ctx, p)
}

func (u *UseCases) DeleteProvider(ctx context.Context, id int64) error {
	return u.repo.DeleteProvider(ctx, id)
}

func (u *UseCases) ArchiveProvider(ctx context.Context, id int64) error {
	return u.repo.ArchiveProvider(ctx, id)
}

func (u *UseCases) RestoreProvider(ctx context.Context, id int64) error {
	return u.repo.RestoreProvider(ctx, id)
}
