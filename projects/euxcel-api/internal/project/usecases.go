package project

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases creates a new instance of Project use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	return u.repo.CreateProject(ctx, p)
}

func (u *useCases) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return u.repo.ListProjects(ctx)
}

func (u *useCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	return u.repo.GetProject(ctx, id)
}

func (u *useCases) UpdateProject(ctx context.Context, p *domain.Project) error {
	return u.repo.UpdateProject(ctx, p)
}

func (u *useCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}
