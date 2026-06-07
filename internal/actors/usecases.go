package actors

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
)

type RepositoryPort interface {
	Resolve(context.Context, domain.ResolveInput) (domain.ResolveResult, error)
	GetByTaxID(context.Context, string) (*domain.Actor, error)
	Search(context.Context, string, string, int) (domain.SearchResult, error)
	List(context.Context, string, int, int) ([]domain.Actor, int64, error)
	Get(context.Context, int64) (*domain.Actor, error)
	Update(context.Context, *domain.Actor) error
	Archive(context.Context, int64) error
	Restore(context.Context, int64) error
	Delete(context.Context, int64) error
	SetRoles(context.Context, int64, []string) error
	SetTaxID(context.Context, int64, string) error
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

func (u *UseCases) Search(ctx context.Context, q string, field string, limit int) (domain.SearchResult, error) {
	return u.repo.Search(ctx, q, field, limit)
}

func (u *UseCases) List(ctx context.Context, status string, page, perPage int) ([]domain.Actor, int64, error) {
	return u.repo.List(ctx, status, page, perPage)
}

func (u *UseCases) Get(ctx context.Context, id int64) (*domain.Actor, error) {
	return u.repo.Get(ctx, id)
}

func (u *UseCases) Update(ctx context.Context, a *domain.Actor) error {
	return u.repo.Update(ctx, a)
}

func (u *UseCases) Archive(ctx context.Context, id int64) error {
	return u.repo.Archive(ctx, id)
}

func (u *UseCases) Restore(ctx context.Context, id int64) error {
	return u.repo.Restore(ctx, id)
}

func (u *UseCases) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

func (u *UseCases) SetRoles(ctx context.Context, id int64, roles []string) error {
	return u.repo.SetRoles(ctx, id, roles)
}

func (u *UseCases) SetTaxID(ctx context.Context, id int64, taxID string) error {
	return u.repo.SetTaxID(ctx, id, taxID)
}
