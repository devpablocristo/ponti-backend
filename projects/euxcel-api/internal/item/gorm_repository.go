package item

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	models "github.com/alphacodinggroup/euxcel-backend/internal/item/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/internal/item/usecases/domain"
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateAssessment(ctx context.Context, item *domain.Item) (string, error) {
	// Validar que item no sea nil.
	if item == nil {
		return "", errors.New("item is nil")
	}

	// Convertir de dominio a modelo.
	model := models.FromDomainAssessment(item)

	// Asignar un nuevo UUID para el ID.
	model.ID = uuid.New().String()

	// Ejecutar la inserción en la base de datos.
	err := r.db.Client().WithContext(ctx).Create(model).Error
	if err != nil {
		return "", fmt.Errorf("failed to create item: %w", err)
	}

	return model.ID, nil
}

func (r *repository) ListAssessments(ctx context.Context) ([]domain.Item, error) {
	var models []models.CreateAssessment
	if err := r.db.Client().WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	assessments := make([]domain.Item, 0, len(models))
	for _, m := range models {
		assessments = append(assessments, *m.ToDomain())
	}
	return assessments, nil
}

func (r *repository) GetAssessment(ctx context.Context, id string) (*domain.Item, error) {
	var model models.Item
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateAssessment(ctx context.Context, item *domain.Item) error {
	model := &models.CreateAssessment{}
	models.FromDomainAssessment(item)
	return r.db.Client().WithContext(ctx).Save(model).Error
}

// DeleteAssessment elimina un paciente a partir de su ID.
func (r *repository) DeleteAssessment(ctx context.Context, id string) error {
	return r.db.Client().WithContext(ctx).Delete(&models.CreateAssessment{}, "id = ?", id).Error
}
