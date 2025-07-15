package labor

import (
	"context"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
)

type RepositoryPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int) ([]domain.ListedLabor, int64, error)
	deleteLabor(context.Context, int64) error
	UpdateLabor(context.Context, *domain.Labor) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	return u.repo.CreateLabor(ctx, labor)
}

func (u *UseCases) ListLabor(ctx context.Context, page, perPage int) ([]domain.ListedLabor, int64, error) {
	return u.repo.ListLabor(ctx, page, perPage)
}

func (u *UseCases) deleteLabor(ctx context.Context, laborId int64) error {
	return u.repo.deleteLabor(ctx, laborId)
}

func (u *UseCases) updateLabor(ctx context.Context, labor *domain.Labor) error {
	return u.repo.UpdateLabor(ctx, labor)
}
