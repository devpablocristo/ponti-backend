package commercialization

import (
	"context"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"

	domain "github.com/devpablocristo/ponti-backend/internal/commercialization/usecases/domain"
)

type RepositoryPort interface {
	CreateBulk(context.Context, []domain.CropCommercialization) error
	ListByProject(context.Context, int64) ([]domain.CropCommercialization, error)
	Update(context.Context, *domain.CropCommercialization) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateOrUpdateBulk(ctx context.Context, items []domain.CropCommercialization) error {
	if len(items) == 0 {
		return domainerr.Validation("no items provided")
	}

	for i := range items {
		items[i].NetPrice = items[i].CalculateNetPrice()
	}

	var toCreate []domain.CropCommercialization
	for _, it := range items {
		if it.ID == 0 {
			toCreate = append(toCreate, it)
		} else {
			if err := u.repo.Update(ctx, &it); err != nil {
				return domainerr.Internal("failed to update crop commercialization")
			}
		}
	}

	if len(toCreate) > 0 {
		if err := u.repo.CreateBulk(ctx, toCreate); err != nil {
			return domainerr.Internal("failed to create crop commercialization")
		}
	}

	return nil
}

func (u *UseCases) ListByProject(ctx context.Context, projectId int64) ([]domain.CropCommercialization, error) {
	return u.repo.ListByProject(ctx, projectId)
}
