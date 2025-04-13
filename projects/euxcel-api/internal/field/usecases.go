package field

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases creates a new instance of Field use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	return u.repo.CreateField(ctx, f)
}

func (u *useCases) ListFields(ctx context.Context) ([]domain.Field, error) {
	return u.repo.ListFields(ctx)
}

func (u *useCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	return u.repo.GetField(ctx, id)
}

func (u *useCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}

func (u *useCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}
