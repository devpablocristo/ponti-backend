package assessment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	models "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/repository/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

func (r *repository) StoreLink(ctx context.Context, link *domain.Link) (string, error) {
	if link == nil {
		return "", errors.New("link is nil")
	}

	model := models.FromDomainToLink(link)
	if model == nil {
		return "", errors.New("failed to convert domain link to model")
	}
	model.ID = uuid.New().String()

	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return "", fmt.Errorf("failed to store link: %w", err)
	}

	return model.ID, nil
}

func (r *repository) GetLink(ctx context.Context, id string) (*domain.Link, error) {
	var model models.Link
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}
