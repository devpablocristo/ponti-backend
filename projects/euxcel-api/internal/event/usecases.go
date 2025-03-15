package event

import (
	"context"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/event/usecases/domain"
)

type useCases struct {
	repository Repository
}

func NewUseCases(r Repository) UseCases {
	return &useCases{
		repository: r,
	}
}

func (u *useCases) ListEvents(ctx context.Context) ([]domain.Event, error) {
	Events, err := u.repository.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	return Events, nil
}

func (u *useCases) CreateEvent(ctx context.Context, Event *domain.Event) error {
	if err := u.repository.CreateEvent(ctx, Event); err != nil {
		return err
	}
	return nil
}

// func (es *useCases) DeleteEvent(ctx context.Context, EventID string) (*Event.domain.Event, error) {
// 	Event, err := es.repo.DeleteEvent(ctx, EventID)
// 	if err != nil {
// 		log.Printf("Error deleting Event with ID %s: %v", EventID, err)
// 		return nil, err
// 	}
// 	return Event, nil
// }

// func (es *useCases) HardDeleteEvent(ctx context.Context, EventID string) (*Event.domain.Event, error) {
// 	Event, err := es.repo.HardDeleteEvent(ctx, EventID)
// 	if err != nil {
// 		log.Printf("Error deleting Event with ID %s: %v", EventID, err)
// 		return nil, err
// 	}
// 	return Event, nil
// }

// func (es *useCases) UpdateEvent(ctx context.Context, Event *Event.domain.Event, EventID string) (*Event.domain.Event, error) {
// 	Event, err := es.repo.UpdateEvent(ctx, Event, EventID)
// 	if err != nil {
// 		log.Printf("Error updating Event with ID %s: %v", EventID, err)
// 		return nil, err
// 	}
// 	return Event, nil
// }

// func (es *useCases) ReviveEvent(ctx context.Context, EventID string) (*Event.domain.Event, error) {
// 	Event, err := es.repo.ReviveEvent(ctx, EventID)
// 	if err != nil {
// 		log.Printf("Error undeleting Event with ID %s: %v", EventID, err)
// 		return nil, err
// 	}
// 	return Event, nil
// }

// func (es *useCases) GetEvent(ctx context.Context, EventID string) (*Event.domain.Event, error) {
// 	Event, err := es.repo.GetEvent(ctx, EventID)
// 	if err != nil {
// 		log.Printf("Error undeleting Event with ID %s: %v", EventID, err)
// 		return nil, err
// 	}
// 	return Event, nil
// }

// func (es *useCases) AddUserToEvent(ctx context.Context, EventID string, user *usr.User) (*Event.domain.Event, error) {
// 	Event, err := es.repo.AddUserToEvent(ctx, EventID, user)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, err
// 	}
// 	return Event, nil
// }
