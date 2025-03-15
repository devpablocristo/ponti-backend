package assessment

import (
	"context"
	"fmt"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

// CreateAssessment crea un nuevo assessment y lo guarda
func (u *useCases) CreateAssessment(ctx context.Context, assessment *domain.Assessment) (string, error) {
	assessmentID, err := u.repository.CreateAssessment(ctx, assessment)
	if err != nil {
		return "", fmt.Errorf("failed to create assessment: %w", err)
	}
	return assessmentID, nil
}

// ListAssessments obtiene la lista de todas las evaluaciones
func (u *useCases) ListAssessments(ctx context.Context) ([]domain.Assessment, error) {
	return u.repository.ListAssessments(ctx)
}

// GetAssessment obtiene una evaluación por su ID
func (u *useCases) GetAssessment(ctx context.Context, assessmentID string) (*domain.Assessment, error) {
	return u.repository.GetAssessment(ctx, assessmentID)
}

// DeleteAssessment elimina una evaluación
func (u *useCases) DeleteAssessment(ctx context.Context, ID string) error {
	return u.repository.DeleteAssessment(ctx, ID)
}

// UpdateAssessment actualiza una evaluación existente
func (u *useCases) UpdateAssessment(ctx context.Context, updateAssessment *domain.Assessment) error {
	return u.repository.UpdateAssessment(ctx, updateAssessment)
}
