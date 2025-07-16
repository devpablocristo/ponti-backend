package unit

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
)

// RepositoryPort is the repository contract for Unit.
type RepositoryPort interface {
	ListUnits(context.Context) ([]domain.Unit, error)
	CreateUnit(context.Context, *domain.Unit) (int64, error)
	UpdateUnit(context.Context, *domain.Unit) error
	DeleteUnit(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) ListUnits(ctx context.Context) ([]domain.Unit, error) {
	return u.repo.ListUnits(ctx)
}
func (u *UseCases) CreateUnit(ctx context.Context, unit *domain.Unit) (int64, error) {
	return u.repo.CreateUnit(ctx, unit)
}
func (u *UseCases) UpdateUnit(ctx context.Context, unit *domain.Unit) error {
	return u.repo.UpdateUnit(ctx, unit)
}
func (u *UseCases) DeleteUnit(ctx context.Context, id int64) error {
	return u.repo.DeleteUnit(ctx, id)
}
