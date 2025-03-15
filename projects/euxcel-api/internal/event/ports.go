package event

import (
	"context"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/event/usecases/domain"
)

type UseCases interface {
	CreateEvent(context.Context, *domain.Event) error
	// DeleteEvent(context.Context, string) (**domain.Event, error)
	// HardDeleteEvent(context.Context, string) (**domain.Event, error)
	// UpdateEvent(context.Context, **domain.Event, string) (**domain.Event, error)
	// ReviveEvent(context.Context, string) (**domain.Event, error)
	// GetEvent(context.Context, string) (**domain.Event, error)
	ListEvents(context.Context) ([]domain.Event, error)
	// AddUserToEvent(context.Context, string, *usr.User) (**domain.Event, error)
}
type Repository interface {
	CreateEvent(context.Context, *domain.Event) error
	// DeleteEvent(context.Context, string) (*domain.Event, error)
	// HardDeleteEvent(context.Context, string) (*domain.Event, error)
	// UpdateEvent(context.Context, *domain.Event, string) (*domain.Event, error)
	// ReviveEvent(context.Context, string) (*domain.Event, error)
	// GetEvent(context.Context, string) (*domain.Event, error)
	ListEvents(context.Context) ([]domain.Event, error)
	// AddUserToEvent(context.Context, string, *usr.User) (*domain.Event, error)
}
