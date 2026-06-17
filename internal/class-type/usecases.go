package classtype

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
)

type RepositoryPort interface {
	CreateClassType(context.Context, *domain.ClassType) (int64, error)
	ListClassTypes(context.Context, string, int, int) ([]domain.ClassType, int64, error)
	GetClassType(context.Context, int64) (*domain.ClassType, error)
	UpdateClassType(context.Context, *domain.ClassType) error
	DeleteClassType(context.Context, int64) error
	ArchiveClassType(context.Context, int64) error
	RestoreClassType(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	return u.repo.CreateClassType(ctx, c)
}

func (u *UseCases) ListClassTypes(ctx context.Context, status string, page, perPage int) ([]domain.ClassType, int64, error) {
	return u.repo.ListClassTypes(ctx, status, page, perPage)
}

func (u *UseCases) GetClassType(ctx context.Context, id int64) (*domain.ClassType, error) {
	return u.repo.GetClassType(ctx, id)
}

func (u *UseCases) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	return u.repo.UpdateClassType(ctx, c)
}

func (u *UseCases) DeleteClassType(ctx context.Context, id int64) error {
	return u.repo.DeleteClassType(ctx, id)
}

func (u *UseCases) ArchiveClassType(ctx context.Context, id int64) error {
	return u.repo.ArchiveClassType(ctx, id)
}

func (u *UseCases) RestoreClassType(ctx context.Context, id int64) error {
	return u.repo.RestoreClassType(ctx, id)
}
