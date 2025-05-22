package project

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type Repository interface {
	CreateProject(ctx context.Context, p *domain.Project) (int64, error)
	ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error)
	ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error)
	GetProject(ctx context.Context, id int64) (*domain.Project, error)
	UpdateProject(ctx context.Context, p *domain.Project) error
	DeleteProject(ctx context.Context, id int64) error
}

type UseCases interface {
	CreateProject(ctx context.Context, p *domain.Project) (int64, error)
	ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error)
	ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error)
	ListProjectsByName(ctx context.Context, name string, page, perPage int) ([]domain.ListedProject, int64, error)
	GetProject(ctx context.Context, id int64) (*domain.Project, error)
	UpdateProject(ctx context.Context, p *domain.Project) error
	DeleteProject(ctx context.Context, id int64) error
}

type WordsSuggester interface {
	ListProjectsByName(ctx context.Context, name string, page, perPage int) ([]domain.ListedProject, int64, error)
}
