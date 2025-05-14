package project

import (
	"context"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	projectdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type UseCases interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	ListProjects(context.Context) ([]domain.Project, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
	ListProjectsByCustomerID(context.Context, int64) ([]projectdom.Project, error)
}

// Repository defines persistence operations for Project.
type Repository interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	ListProjects(context.Context) ([]domain.Project, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
	ListProjectsByCustomerID(context.Context, int64) ([]projectdom.Project, error)
}
