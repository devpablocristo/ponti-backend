package assessment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	models "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/repository/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

func (r *repository) CreateAssessment(ctx context.Context, assessment *domain.Assessment) (string, error) {
	// Validar que assessment no sea nil.
	if assessment == nil {
		return "", errors.New("assessment is nil")
	}

	// Convertir de dominio a modelo.
	model := models.FromDomainAssessment(assessment)

	// Asignar un nuevo UUID para el ID.
	model.ID = uuid.New().String()

	// Ejecutar la inserción en la base de datos.
	err := r.db.Client().WithContext(ctx).Create(model).Error
	if err != nil {
		return "", fmt.Errorf("failed to create assessment: %w", err)
	}

	return model.ID, nil
}

func (r *repository) ListAssessments(ctx context.Context) ([]domain.Assessment, error) {
	var models []models.CreateAssessment
	if err := r.db.Client().WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	assessments := make([]domain.Assessment, 0, len(models))
	for _, m := range models {
		assessments = append(assessments, *m.ToDomain())
	}
	return assessments, nil
}

func (r *repository) GetAssessment(ctx context.Context, id string) (*domain.Assessment, error) {
	var model models.Assessment
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateAssessment(ctx context.Context, assessment *domain.Assessment) error {
	model := &models.CreateAssessment{}
	models.FromDomainAssessment(assessment)
	return r.db.Client().WithContext(ctx).Save(model).Error
}

// DeleteAssessment elimina un paciente a partir de su ID.
func (r *repository) DeleteAssessment(ctx context.Context, id string) error {
	return r.db.Client().WithContext(ctx).Delete(&models.CreateAssessment{}, "id = ?", id).Error
}
