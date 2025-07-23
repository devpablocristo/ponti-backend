package commercialization

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/commercialization/usecases/domain"
)

type RepositoryPort interface {
	CreateBulk(context.Context, []domain.CropCommercialization) error
	ListByProject(context.Context, int64) ([]domain.CropCommercialization, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateBulk(ctx context.Context, items []domain.CropCommercialization) error {
	if len(items) == 0 {
		return types.NewError(types.ErrInvalidInput, "no items provided", nil)
	}
	return u.repo.CreateBulk(ctx, items)
}

func (u *UseCases) ListByProject(ctx context.Context, projectId int64) ([]domain.CropCommercialization, error) {
	return u.repo.ListByProject(ctx, projectId)
}
