package unit

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/usecases/domain"
)

// UnitRepositoryPort is the repository contract for Unit.
type UnitRepositoryPort interface {
	ListUnits(context.Context) ([]domain.Unit, error)
	CreateUnit(context.Context, *domain.Unit) (int64, error)
	UpdateUnit(context.Context, *domain.Unit) error
	DeleteUnit(context.Context, int64) error
}

type UnitUseCases struct {
	repo UnitRepositoryPort
}

func NewUnitUseCases(repo UnitRepositoryPort) *UnitUseCases {
	return &UnitUseCases{repo: repo}
}

func (u *UnitUseCases) ListUnits(ctx context.Context) ([]domain.Unit, error) {
	return u.repo.ListUnits(ctx)
}
func (u *UnitUseCases) CreateUnit(ctx context.Context, unit *domain.Unit) (int64, error) {
	return u.repo.CreateUnit(ctx, unit)
}
func (u *UnitUseCases) UpdateUnit(ctx context.Context, unit *domain.Unit) error {
	return u.repo.UpdateUnit(ctx, unit)
}
func (u *UnitUseCases) DeleteUnit(ctx context.Context, id int64) error {
	return u.repo.DeleteUnit(ctx, id)
}
