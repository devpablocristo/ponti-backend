package manager

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/manager/usecases/domain"
)

type RepositoryPort interface {
	CreateManager(context.Context, *domain.Manager) (int64, error)
	ListManagers(context.Context, int, int) ([]domain.Manager, int64, error)
	GetManager(context.Context, int64) (*domain.Manager, error)
	UpdateManager(context.Context, *domain.Manager) error
	DeleteManager(context.Context, int64) error
	ArchiveManager(context.Context, int64) error
	RestoreManager(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

// NewUseCases crea una instancia de los casos de uso para Manager.
func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateManager(ctx context.Context, m *domain.Manager) (int64, error) {
	return u.repo.CreateManager(ctx, m)
}

func (u *UseCases) ListManagers(ctx context.Context, page, perPage int) ([]domain.Manager, int64, error) {
	return u.repo.ListManagers(ctx, page, perPage)
}

func (u *UseCases) GetManager(ctx context.Context, id int64) (*domain.Manager, error) {
	return u.repo.GetManager(ctx, id)
}

func (u *UseCases) UpdateManager(ctx context.Context, m *domain.Manager) error {
	return u.repo.UpdateManager(ctx, m)
}

func (u *UseCases) DeleteManager(ctx context.Context, id int64) error {
	return u.repo.DeleteManager(ctx, id)
}

func (u *UseCases) ArchiveManager(ctx context.Context, id int64) error {
	return u.repo.ArchiveManager(ctx, id)
}

func (u *UseCases) RestoreManager(ctx context.Context, id int64) error {
	return u.repo.RestoreManager(ctx, id)
}
