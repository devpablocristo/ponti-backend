package item

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/internal/item/usecases/domain"
)

// UseCases define la interfaz pública con los métodos que expondremos
type UseCases interface {
	CreateAssessment(context.Context, *domain.Item) (string, error)
	ListAssessments(context.Context) ([]domain.Item, error)
	GetAssessment(context.Context, string) (*domain.Item, error)
	DeleteAssessment(context.Context, string) error
	UpdateAssessment(context.Context, *domain.Item) error
}

type Repository interface {
	CreateAssessment(context.Context, *domain.Item) (string, error)
	UpdateAssessment(context.Context, *domain.Item) error
	GetAssessment(context.Context, string) (*domain.Item, error)
	DeleteAssessment(context.Context, string) error
	ListAssessments(context.Context) ([]domain.Item, error)
}
