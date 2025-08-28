package workorder

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type RepositoryPort interface {
	CreateWorkorder(context.Context, *domain.Workorder) (int64, error)
	GetWorkorderByID(ctx context.Context, id int64) (*domain.Workorder, error)
	//DuplicateWorkorder(context.Context, string) (string, error)
	UpdateWorkorderByID(context.Context, *domain.Workorder) error
	DeleteWorkorderByID(context.Context, int64) error
	ListWorkorders(context.Context, domain.WorkorderFilter, types.Input) ([]domain.WorkorderListElement, types.PageInfo, error)
	GetMetrics(context.Context, domain.WorkorderFilter) (*domain.WorkorderMetrics, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(r RepositoryPort) *UseCases {
	return &UseCases{repo: r}
}

func (u *UseCases) CreateWorkorder(ctx context.Context, o *domain.Workorder) (int64, error) {
	return u.repo.CreateWorkorder(ctx, o)
}

func (u *UseCases) GetWorkorderByID(ctx context.Context, id int64) (*domain.Workorder, error) {
	return u.repo.GetWorkorderByID(ctx, id)
}

func (u *UseCases) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
	return "", nil
}

func (u *UseCases) UpdateWorkorderByID(ctx context.Context, o *domain.Workorder) error {
	return u.repo.UpdateWorkorderByID(ctx, o)
}

func (u *UseCases) DeleteWorkorderByID(ctx context.Context, id int64) error {
	return u.repo.DeleteWorkorderByID(ctx, id)
}

// ListWorkorders delega al repositorio
func (u *UseCases) ListWorkorders(
	ctx context.Context,
	filt domain.WorkorderFilter,
	inp types.Input,
) ([]domain.WorkorderListElement, types.PageInfo, error) {
	return u.repo.ListWorkorders(ctx, filt, inp)
}

// GetMetrics delega al repositorio
func (u *UseCases) GetMetrics(ctx context.Context, f domain.WorkorderFilter) (*domain.WorkorderMetrics, error) {
	return u.repo.GetMetrics(ctx, f)
}
