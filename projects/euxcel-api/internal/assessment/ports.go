package assessment

import (
	"context"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

// UseCases define la interfaz pública con los métodos que expondremos
type UseCases interface {
	CreateAssessment(context.Context, *domain.Assessment) (string, error)
	ListAssessments(context.Context) ([]domain.Assessment, error)
	GetAssessment(context.Context, string) (*domain.Assessment, error)
	DeleteAssessment(context.Context, string) error
	UpdateAssessment(context.Context, *domain.Assessment) error

	// INFO: Assessment Link
	GenerateLink(context.Context, string) (string, error)
	SendLink(context.Context, string) error
	GetLink(context.Context, string) (*domain.Link, error)
	ValidateLink(context.Context, string) (*domain.Link, error)
}

type Repository interface {
	CreateAssessment(context.Context, *domain.Assessment) (string, error)
	UpdateAssessment(context.Context, *domain.Assessment) error
	GetAssessment(context.Context, string) (*domain.Assessment, error)
	DeleteAssessment(context.Context, string) error
	ListAssessments(context.Context) ([]domain.Assessment, error)

	// INFO: Assessment Link
	StoreLink(context.Context, *domain.Link) (string, error)
	GetLink(context.Context, string) (*domain.Link, error)
}
