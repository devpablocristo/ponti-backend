package group

import (
	"context"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/group/usecases/domain"
)

type Repository interface {
	CreateGroup(context.Context, *domain.Group) error
	GetGroupByID(context.Context, uint) (*domain.Group, error)
	UpdateGroup(context.Context, *domain.Group) error
	DeleteGroup(context.Context, uint) error
	ListGroups(context.Context) ([]domain.Group, error)
}

type UseCases interface {
	CreateGroup(context.Context, *domain.Group) error
	GetGroupByID(context.Context, uint) (*domain.Group, error)
	UpdateGroup(context.Context, *domain.Group) error
	DeleteGroup(context.Context, uint) error
	ListGroups(context.Context) ([]domain.Group, error)
}
