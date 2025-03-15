package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/event/usecases/domain"
)

type Event struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Location    string             `bson:"location" json:"location"`
	StartTime   time.Time          `bson:"start_time" json:"start_time"`
	EndTime     time.Time          `bson:"end_time" json:"end_time"`
	Category    string             `bson:"category" json:"category"`
	CreatorID   primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	IsPublic    bool               `bson:"is_public" json:"is_public"`
	IsRecurring bool               `bson:"is_recurring" json:"is_recurring"`
	SeriesID    primitive.ObjectID `bson:"series_id,omitempty" json:"series_id"`
	Status      string             `bson:"status" json:"status"`
	Organizer   []string           `bson:"organizer" json:"organizer"`
	Attendees   []string           `bson:"attendees" json:"attendees"`
	Planners    []string           `bson:"planners" json:"planners"`
	Tags        []string           `bson:"tags" json:"tags"`
	CreatedAt   time.Time          `bson:"create_at" json:"create_at"`
}

func (m *Event) ToDomain() *domain.Event {
	return &domain.Event{
		ID:          m.ID.Hex(),
		Title:       m.Title,
		Description: m.Description,
		Location:    m.Location,
		StartTime:   m.StartTime,
		EndTime:     m.EndTime,
		Category:    domain.Category(m.Category),
		CreatorID:   m.CreatorID.Hex(),
		IsPublic:    m.IsPublic,
		IsRecurring: m.IsRecurring,
		SeriesID:    m.SeriesID.Hex(),
		Status:      domain.EventStatus(m.Status),
		Organizers:  m.Organizer,
		Attendees:   m.Attendees,
		Planners:    m.Planners,
		Tags:        m.Tags,
		CreatedAt:   m.CreatedAt,
	}
}

// FromDomain convierte un Event del dominio en un Event.
// Retorna un error si alguna conversión de ID falla.
func (m *Event) FromDomain(event *domain.Event) (*Event, error) {
	var err error

	// Convertir ID
	if event.ID != "" {
		m.ID, err = primitive.ObjectIDFromHex(event.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid ID format: %w", err)
		}
	} else {
		// Si no se proporciona un ID, se puede generar uno nuevo.
		m.ID = primitive.NewObjectID()
	}

	// Convertir CreatorID
	if event.CreatorID != "" {
		m.CreatorID, err = primitive.ObjectIDFromHex(event.CreatorID)
		if err != nil {
			return nil, fmt.Errorf("invalid CreatorID format: %w", err)
		}
	}

	// Convertir SeriesID
	if event.SeriesID != "" {
		m.SeriesID, err = primitive.ObjectIDFromHex(event.SeriesID)
		if err != nil {
			return nil, fmt.Errorf("invalid SeriesID format: %w", err)
		}
	}

	// Asignar otros campos
	m.Title = event.Title
	m.Description = event.Description
	m.Location = event.Location
	m.StartTime = event.StartTime
	m.EndTime = event.EndTime
	m.Category = string(event.Category)
	m.IsPublic = event.IsPublic
	m.IsRecurring = event.IsRecurring
	m.Status = string(event.Status)
	m.Organizer = event.Organizers
	m.Attendees = event.Attendees
	m.Planners = event.Planners
	m.Tags = event.Tags
	m.CreatedAt = event.CreatedAt

	return m, nil
}

type EventList []Event

func (list *EventList) ToDomain() []domain.Event {
	events := make([]domain.Event, len(*list))
	for i, model := range *list {
		events[i] = *model.ToDomain()
	}
	return events
}
