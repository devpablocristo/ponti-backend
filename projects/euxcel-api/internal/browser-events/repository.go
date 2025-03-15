package browserEvent

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	mng "github.com/alphacodinggroup/euxcel-backend/pkg/databases/nosql/mongodb/mongo-driver"

	"github.com/alphacodinggroup/euxcel-backend/internal/browser-events/repository/models"
	"github.com/alphacodinggroup/euxcel-backend/internal/browser-events/usecases/domain"
)

type mongoRepository struct {
	repository mng.Repository
}

func NewRepository(r mng.Repository) Repository {
	return &mongoRepository{
		repository: r,
	}
}

func (r *mongoRepository) SaveBrowserEvent(ctx context.Context, event *domain.BrowserEvent) error {
	// Conversión de la entidad de dominio a nuestro modelo para Mongo.
	m, err := models.FromDomain(event)
	if err != nil {
		return err
	}

	// Asignar la fecha de creación.
	m.CreatedAt = time.Now()

	// Insertar el documento en la colección "browser_events".
	_, err = r.repository.DB().Collection("browser_events").InsertOne(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

// GetBrowserEventsByCandidateID retorna todos los eventos asociados a un CandidateID.
func (r *mongoRepository) GetBrowserEventsByCandidateID(ctx context.Context, candidateID string) ([]*domain.BrowserEvent, error) {
	// Filtro basado en el campo "candidate_id".
	filter := bson.M{"candidate_id": candidateID}

	cursor, err := r.repository.DB().Collection("browser_events").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.BrowserEvent
	for cursor.Next(ctx) {
		var m models.BrowserEvent
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		event, err := m.ToDomain()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// GetBrowserEventsByAsssementID retorna todos los eventos asociados a un AssessmentID.
// Ten en cuenta que se asume que en el modelo, "assessment_ids" es un slice/array y
// Mongo realiza la búsqueda en el array de forma automática.
func (r *mongoRepository) GetBrowserEventsByAsssementID(ctx context.Context, assessmentID string) ([]*domain.BrowserEvent, error) {
	// Filtro basado en el campo "assessment_ids".
	filter := bson.M{"assessment_ids": assessmentID}

	cursor, err := r.repository.DB().Collection("browser_events").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []*domain.BrowserEvent
	for cursor.Next(ctx) {
		var m models.BrowserEvent
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		event, err := m.ToDomain()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
