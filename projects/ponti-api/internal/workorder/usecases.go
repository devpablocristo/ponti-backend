package workorder

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkorder(context.Context, *domain.Workorder) (string, error)
	GetWorkorder(context.Context, string) (*domain.Workorder, error)
	DuplicateWorkorder(context.Context, string) (string, error)
	UpdateWorkorder(context.Context, *domain.Workorder) error
	DeleteWorkorder(context.Context, string) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(r RepositoryPort) *UseCases {
	return &UseCases{repo: r}
}

func (u *UseCases) CreateWorkorder(ctx context.Context, o *domain.Workorder) (string, error) {
	return u.repo.CreateWorkorder(ctx, o)
}

func (u *UseCases) GetWorkorder(ctx context.Context, number string) (*domain.Workorder, error) {
	return u.repo.GetWorkorder(ctx, number)
}

func (u *UseCases) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
	return u.repo.DuplicateWorkorder(ctx, number)
}

func (u *UseCases) UpdateWorkorder(ctx context.Context, o *domain.Workorder) error {
	return u.repo.UpdateWorkorder(ctx, o)
}

func (u *UseCases) DeleteWorkorder(ctx context.Context, number string) error {
	return u.repo.DeleteWorkorder(ctx, number)
}
