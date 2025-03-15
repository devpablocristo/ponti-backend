package person

import (
	"context"

	"github.com/devpablocristo/monorepo/projects/qh/internal/person/usecases/domain"
)

type UseCases interface {
	CreatePerson(context.Context, *domain.Person) (string, error)
	ListPersons(context.Context) ([]domain.Person, error)
	GetPerson(context.Context, string) (*domain.Person, error)
	UpdatePerson(context.Context, string, *domain.Person) error
	DeletePerson(context.Context, string, bool) error
}

type Repository interface {
	CreatePerson(context.Context, *domain.Person) (string, error)
	ListPersons(context.Context) ([]domain.Person, error)
	GetPerson(context.Context, string) (*domain.Person, error)
	UpdatePerson(context.Context, string, *domain.Person) error
	DeletePerson(context.Context, string, bool) error
}
