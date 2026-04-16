package field

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
)

type RepositoryPort interface {
	CreateField(context.Context, *domain.Field) (int64, error)
	ListFields(context.Context, int, int) ([]domain.Field, int64, error)
	GetField(context.Context, int64) (*domain.Field, error)
	UpdateField(context.Context, *domain.Field) error
	DeleteField(context.Context, int64) error
	ArchiveField(context.Context, int64) error
	RestoreField(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	return u.repo.CreateField(ctx, f)
}

func (u *UseCases) ListFields(ctx context.Context, page, perPage int) ([]domain.Field, int64, error) {
	return u.repo.ListFields(ctx, page, perPage)
}

func (u *UseCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	return u.repo.GetField(ctx, id)
}

func (u *UseCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}

func (u *UseCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}

func (u *UseCases) ArchiveField(ctx context.Context, id int64) error {
	return u.repo.ArchiveField(ctx, id)
}

func (u *UseCases) RestoreField(ctx context.Context, id int64) error {
	return u.repo.RestoreField(ctx, id)
}
