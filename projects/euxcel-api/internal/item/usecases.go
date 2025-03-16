package item

import (
	"context"
	"fmt"

	authe "github.com/alphacodinggroup/euxcel-backend/internal/authe"
	candidate "github.com/alphacodinggroup/euxcel-backend/internal/candidate"
	config "github.com/alphacodinggroup/euxcel-backend/internal/config"
	domain "github.com/alphacodinggroup/euxcel-backend/internal/item/usecases/domain"
	notification "github.com/alphacodinggroup/euxcel-backend/internal/notification"
	person "github.com/alphacodinggroup/euxcel-backend/internal/person"
)

// useCases implementa la interfaz UseCases
type useCases struct {
	repository     Repository
	config         config.Loader
	autheUc        authe.UseCases
	candidateUc    candidate.UseCases
	personUc       person.UseCases
	notificationUc notification.UseCases
}

// NewUseCases crea una instancia de useCases con las dependencias adecuadas
func NewUseCases(
	repo Repository,
	notif notification.UseCases,
	candidate candidate.UseCases,
	cfg config.Loader,
	au authe.UseCases,
	pn person.UseCases,
) UseCases {
	return &useCases{
		repository:     repo,
		notificationUc: notif,
		candidateUc:    candidate,
		config:         cfg,
		autheUc:        au,
		personUc:       pn,
	}
}

// CreateAssessment crea un nuevo item y lo guarda
func (u *useCases) CreateAssessment(ctx context.Context, item *domain.Item) (string, error) {
	assessmentID, err := u.repository.CreateAssessment(ctx, item)
	if err != nil {
		return "", fmt.Errorf("failed to create item: %w", err)
	}
	return assessmentID, nil
}

// ListAssessments obtiene la lista de todas las evaluaciones
func (u *useCases) ListAssessments(ctx context.Context) ([]domain.Item, error) {
	return u.repository.ListAssessments(ctx)
}

// GetAssessment obtiene una evaluación por su ID
func (u *useCases) GetAssessment(ctx context.Context, assessmentID string) (*domain.Item, error) {
	return u.repository.GetAssessment(ctx, assessmentID)
}

// DeleteAssessment elimina una evaluación
func (u *useCases) DeleteAssessment(ctx context.Context, ID string) error {
	return u.repository.DeleteAssessment(ctx, ID)
}

// UpdateAssessment actualiza una evaluación existente
func (u *useCases) UpdateAssessment(ctx context.Context, updateAssessment *domain.Item) error {
	return u.repository.UpdateAssessment(ctx, updateAssessment)
}
