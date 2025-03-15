package group

import (
	"context"

	"github.com/devpablocristo/monorepo/projects/qh/internal/group/usecases/domain"
)

type useCases struct {
	repository Repository
}

func NewUseCases(repository Repository) UseCases {
	return &useCases{
		repository: repository,
	}
}

// CreateGroup maneja la creación de un nuevo grupo
func (u *useCases) CreateGroup(ctx context.Context, group *domain.Group) error {
	return u.repository.CreateGroup(ctx, group)
}

// GetGroupByID obtiene un grupo por su ID
func (u *useCases) GetGroupByID(ctx context.Context, id uint) (*domain.Group, error) {
	return u.repository.GetGroupByID(ctx, id)
}

// UpdateGroup actualiza la información de un grupo existente
func (u *useCases) UpdateGroup(ctx context.Context, group *domain.Group) error {
	return u.repository.UpdateGroup(ctx, group)
}

// DeleteGroup elimina un grupo por su ID
func (u *useCases) DeleteGroup(ctx context.Context, id uint) error {
	return u.repository.DeleteGroup(ctx, id)
}

// ListGroups lista todos los grupos disponibles
func (u *useCases) ListGroups(ctx context.Context) ([]domain.Group, error) {
	return u.repository.ListGroups(ctx)
}
