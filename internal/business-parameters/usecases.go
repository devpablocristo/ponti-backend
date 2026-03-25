package bparams

import (
	"context"

	"github.com/devpablocristo/core/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
)

type RepositoryPort interface {
	GetByKey(ctx context.Context, key string) (*domain.BusinessParameter, error)
	ListByCategory(ctx context.Context, category string) ([]domain.BusinessParameter, error)
	ListAll(ctx context.Context) ([]domain.BusinessParameter, error)
	Create(ctx context.Context, item *domain.BusinessParameter) (int64, error)
	Update(ctx context.Context, item *domain.BusinessParameter) error
	Delete(ctx context.Context, id int64) error
}

type UseCases struct {
	repository RepositoryPort
}

func NewUseCases(repository RepositoryPort) *UseCases {
	return &UseCases{
		repository: repository,
	}
}

func (u *UseCases) GetParameter(ctx context.Context, key string) (*domain.BusinessParameter, error) {
	if key == "" {
		return nil, domainerr.Validation("key is required")
	}

	return u.repository.GetByKey(ctx, key)
}

func (u *UseCases) GetParametersByCategory(ctx context.Context, category string) ([]domain.BusinessParameter, error) {
	if category == "" {
		return nil, domainerr.Validation("category is required")
	}

	return u.repository.ListByCategory(ctx, category)
}

func (u *UseCases) GetAllParameters(ctx context.Context) ([]domain.BusinessParameter, error) {
	return u.repository.ListAll(ctx)
}

func (u *UseCases) CreateParameter(ctx context.Context, param *domain.BusinessParameter) (int64, error) {
	if param.Key == "" || param.Value == "" || param.Type == "" || param.Category == "" {
		return 0, domainerr.Validation("missing required fields")
	}

	return u.repository.Create(ctx, param)
}

func (u *UseCases) UpdateParameter(ctx context.Context, param *domain.BusinessParameter) error {
	if param.ID == 0 {
		return domainerr.Validation("invalid id")
	}

	if param.Key == "" || param.Value == "" || param.Type == "" || param.Category == "" {
		return domainerr.Validation("missing required fields")
	}

	return u.repository.Update(ctx, param)
}

func (u *UseCases) DeleteParameter(ctx context.Context, id int64) error {
	if id == 0 {
		return domainerr.Validation("invalid id")
	}

	return u.repository.Delete(ctx, id)
}
