package provider

import (
	"context"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
)

type RepositoryPort interface {
	GetProviders(context.Context) ([]domain.Provider, error)
	ListArchivedProviders(context.Context) ([]domain.Provider, error)
	GetProvider(context.Context, int64) (*domain.Provider, error)
	CreateProvider(context.Context, *domain.Provider) (int64, error)
	UpdateProvider(context.Context, *domain.Provider) error
	ArchiveProvider(context.Context, int64) error
	RestoreProvider(context.Context, int64) error
	HardDeleteProvider(context.Context, int64) error
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

func (u *UseCases) ListArchivedProviders(ctx context.Context) ([]domain.Provider, error) {
	return u.repo.ListArchivedProviders(ctx)
}

func (u *UseCases) GetProvider(ctx context.Context, id int64) (*domain.Provider, error) {
	if id == 0 {
		return nil, domainerr.Validation("invalid id")
	}
	return u.repo.GetProvider(ctx, id)
}

func (u *UseCases) CreateProvider(ctx context.Context, provider *domain.Provider) (int64, error) {
	if provider == nil || provider.Name == "" {
		return 0, domainerr.Validation("provider name is required")
	}
	return u.repo.CreateProvider(ctx, provider)
}

func (u *UseCases) UpdateProvider(ctx context.Context, provider *domain.Provider) error {
	if provider == nil || provider.ID == 0 {
		return domainerr.Validation("invalid id")
	}
	if provider.Name == "" {
		return domainerr.Validation("provider name is required")
	}
	return u.repo.UpdateProvider(ctx, provider)
}

func (u *UseCases) ArchiveProvider(ctx context.Context, id int64) error {
	if id == 0 {
		return domainerr.Validation("invalid id")
	}
	return u.repo.ArchiveProvider(ctx, id)
}

func (u *UseCases) RestoreProvider(ctx context.Context, id int64) error {
	if id == 0 {
		return domainerr.Validation("invalid id")
	}
	return u.repo.RestoreProvider(ctx, id)
}

func (u *UseCases) HardDeleteProvider(ctx context.Context, id int64) error {
	if id == 0 {
		return domainerr.Validation("invalid id")
	}
	return u.repo.HardDeleteProvider(ctx, id)
}

