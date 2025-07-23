package workorder

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkOrder(context.Context, *domain.WorkOrder) (string, error)
	GetWorkOrder(context.Context, string) (*domain.WorkOrder, error)
	DuplicateWorkOrder(context.Context, string) (string, error)
	UpdateWorkOrder(context.Context, *domain.WorkOrder) error
	DeleteWorkOrder(context.Context, string) error
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

func (u *UseCases) GetWorkOrder(ctx context.Context, number string) (*domain.WorkOrder, error) {
	return u.repo.GetWorkOrder(ctx, number)
}

func (u *UseCases) DuplicateWorkOrder(ctx context.Context, number string) (string, error) {
	return u.repo.DuplicateWorkOrder(ctx, number)
}

func (u *UseCases) UpdateWorkOrder(ctx context.Context, o *domain.WorkOrder) error {
	return u.repo.UpdateWorkOrder(ctx, o)
}

func (u *UseCases) DeleteWorkOrder(ctx context.Context, number string) error {
	return u.repo.DeleteWorkOrder(ctx, number)
}
