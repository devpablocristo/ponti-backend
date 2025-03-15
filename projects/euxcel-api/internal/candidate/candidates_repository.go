package candidate

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	gorm "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"

	models "github.com/devpablocristo/monorepo/projects/qh/internal/candidate/repository/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/candidate/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateCandidate(ctx context.Context, candidate *domain.Candidate) (string, error) {
	if candidate == nil {
		return "", errors.New("candidate is nil")
	}

	model, err := models.FromDomainCandidate(candidate)
	if err != nil {
		return "", err
	}
	model.ID = uuid.New().String()

	err = r.db.Client().WithContext(ctx).Create(model).Error
	if err != nil {
		return "", fmt.Errorf("failed to create candidate: %w", err)
	}

	return model.ID, nil
}

func (r *repository) ListCandidates(ctx context.Context) ([]domain.Candidate, error) {
	var models []models.CreateCandidate
	if err := r.db.Client().WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	users := make([]domain.Candidate, 0, len(models))
	for _, m := range models {
		candidate, err := m.ToDomain()
		if err != nil {
			return nil, err
		}
		users = append(users, *candidate)
	}
	return users, nil
}

func (r *repository) GetCandidate(ctx context.Context, id string) (*domain.Candidate, error) {
	var model models.Candidate
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	candidate, err := model.ToDomain()
	if err != nil {
		return nil, err
	}
	return candidate, nil
}

func (r *repository) UpdateCandidate(ctx context.Context, candidate *domain.Candidate) error {
	model := &models.CreateCandidate{}
	models.FromDomainCandidate(candidate)
	return r.db.Client().WithContext(ctx).Save(model).Error
}

func (r *repository) DeleteCandidate(ctx context.Context, id string) error {
	return r.db.Client().WithContext(ctx).Delete(&models.CreateCandidate{}, "id = ?", id).Error
}
