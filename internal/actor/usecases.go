package actor

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
)

type RepositoryPort interface {
	CreateActor(context.Context, *domain.Actor) (int64, error)
	ListActors(context.Context, domain.ListFilters, int, int) ([]domain.Actor, int64, error)
	GetActor(context.Context, int64) (*domain.Actor, error)
	UpdateActor(context.Context, *domain.Actor) error
	ArchiveActor(context.Context, int64) error
	RestoreActor(context.Context, int64) error
	HardDeleteActor(context.Context, int64) error
	AddRole(context.Context, int64, string) error
	AddAlias(context.Context, int64, domain.ActorAlias) (int64, error)
	MergeActors(context.Context, domain.MergeRequest) (*domain.MergeImpact, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateActor(ctx context.Context, actor *domain.Actor) (int64, error) {
	return u.repo.CreateActor(ctx, actor)
}

func (u *UseCases) ListActors(ctx context.Context, filters domain.ListFilters, page, perPage int) ([]domain.Actor, int64, error) {
	return u.repo.ListActors(ctx, filters, page, perPage)
}

func (u *UseCases) GetActor(ctx context.Context, id int64) (*domain.Actor, error) {
	return u.repo.GetActor(ctx, id)
}

func (u *UseCases) UpdateActor(ctx context.Context, actor *domain.Actor) error {
	return u.repo.UpdateActor(ctx, actor)
}

func (u *UseCases) ArchiveActor(ctx context.Context, id int64) error {
	return u.repo.ArchiveActor(ctx, id)
}

func (u *UseCases) RestoreActor(ctx context.Context, id int64) error {
	return u.repo.RestoreActor(ctx, id)
}

func (u *UseCases) HardDeleteActor(ctx context.Context, id int64) error {
	return u.repo.HardDeleteActor(ctx, id)
}

func (u *UseCases) AddRole(ctx context.Context, id int64, role string) error {
	return u.repo.AddRole(ctx, id, role)
}

func (u *UseCases) AddAlias(ctx context.Context, id int64, alias domain.ActorAlias) (int64, error) {
	return u.repo.AddAlias(ctx, id, alias)
}

func (u *UseCases) MergeActors(ctx context.Context, req domain.MergeRequest) (*domain.MergeImpact, error) {
	return u.repo.MergeActors(ctx, req)
}
