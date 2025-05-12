package project

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

// UseCases defines business operations for Project.
type UseCases interface {
	CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error)
	ListProjects(ctx context.Context) ([]domain.Project, error)
	GetProject(ctx context.Context, id int64) (*domain.Project, error)
	// UpdateProject(ctx context.Context, p *domain.Project) error
	DeleteProject(ctx context.Context, id int64) error
}

// Repository defines operations for Project.
type Repository interface {
	CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error)
	ListProjects(ctx context.Context) ([]domain.Project, error)
	GetProject(ctx context.Context, id int64) (*domain.Project, error)
	// UpdateProject(ctx context.Context, p *domain.Project) error
	DeleteProject(ctx context.Context, id int64) error
}
