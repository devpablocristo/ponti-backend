package workorder

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (string, error)
	GetOrder(context.Context, string) (*domain.WorkOrder, error)
	DuplicateOrder(context.Context, string) (string, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(r RepositoryPort) *UseCases {
	return &UseCases{repo: r}
}

func (u *UseCases) CreateWorkOrder(ctx context.Context, o *domain.WorkOrder) (string, error) {
	return u.repo.CreateWorkOrder(ctx, o)
}

func (u *UseCases) GetOrder(ctx context.Context, number string) (*domain.WorkOrder, error) {
	return u.repo.GetOrder(ctx, number)
}

func (u *UseCases) DuplicateOrder(ctx context.Context, number string) (string, error) {
	return u.repo.DuplicateOrder(ctx, number)
}
