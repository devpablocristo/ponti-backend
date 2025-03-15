package dto

import (
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/event/usecases/domain"
)

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Category    string    `json:"category"`
	CreatorID   string    `json:"creator_id"`
	IsPublic    bool      `json:"is_public"`
	IsRecurring bool      `json:"is_recurring"`
	SeriesID    string    `json:"series_id"`
	Status      string    `json:"status"`
	Organizer   []string  `json:"organizer"`
	Attendees   []string  `json:"attendees"`
	Planners    []string  `json:"planners"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"create_at"`
}

// ToDomain convierte un BaseEvent en un Event del dominio
func (dto *Event) ToDomain() *domain.Event {
	// Retornar el objeto del dominio con los datos del DTO
	return &domain.Event{
		ID:          dto.ID,
		Title:       dto.Title,
		Description: dto.Description,
		Location:    dto.Location,
		StartTime:   dto.StartTime,
		EndTime:     dto.EndTime,
		Category:    domain.Category(dto.Category),
		CreatorID:   dto.CreatorID,
		IsPublic:    dto.IsPublic,
		IsRecurring: dto.IsRecurring,
		SeriesID:    dto.SeriesID,
		Status:      domain.EventStatus(dto.Status),
		Organizers:  dto.Organizer,
		Attendees:   dto.Attendees,
		Planners:    dto.Planners,
		Tags:        dto.Tags,
		CreatedAt:   dto.CreatedAt,
	}
}

func (dto *Event) FromDomain(event *domain.Event) *Event {
	dto.ID = event.ID
	dto.Title = event.Title
	dto.Description = event.Description
	dto.Location = event.Location
	dto.StartTime = event.StartTime
	dto.EndTime = event.EndTime
	dto.Category = string(event.Category)
	dto.CreatorID = event.CreatorID
	dto.IsPublic = event.IsPublic
	dto.IsRecurring = event.IsRecurring
	dto.SeriesID = event.SeriesID
	dto.Status = string(event.Status)
	dto.Organizer = event.Organizers
	dto.Attendees = event.Attendees
	dto.Planners = event.Planners
	dto.Tags = event.Tags
	dto.CreatedAt = event.CreatedAt

	return dto
}

type EventList []Event

func (list *EventList) ToDomain() []domain.Event {
	events := make([]domain.Event, len(*list))
	for i, event := range *list {
		events[i] = *event.ToDomain()
	}
	return events
}

// FromDomain convierte una lista de domain.Event en una EventList y la retorna
func (list *EventList) FromDomain(events []domain.Event) EventList {
	*list = make(EventList, len(events))
	for i, event := range events {
		var dto Event
		dto.FromDomain(&event)
		(*list)[i] = dto
	}
	return *list
}
