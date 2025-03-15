package models

import (
	"errors"
	"time"

	"github.com/alphacodinggroup/euxcel-backend/internal/browser-events/usecases/domain"
)

// BrowserEvent representa el evento para almacenamiento en MongoDB.
type BrowserEvent struct {
	// ID es el identificador único asignado por Mongo (se mapea a _id).
	ID string `json:"id,omitempty" bson:"_id,omitempty"`

	// EventType indica el tipo de evento (por ejemplo, "click", "keydown", etc.).
	EventType string `json:"eventType" bson:"eventType"`

	// CandidateID identifica al candidato relacionado con el evento.
	CandidateID string `json:"candidateId" bson:"candidateId"`

	// AssessmentIDs contiene los IDs de las evaluaciones asociadas.
	AssessmentIDs []string `json:"assessmentIds,omitempty" bson:"assessmentIds,omitempty"`

	// Timestamp marca el momento en que ocurrió el evento.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`

	// TargetID es el identificador del elemento HTML sobre el que se produjo el evento.
	TargetID string `json:"targetId,omitempty" bson:"targetId,omitempty"`

	// Payload contiene datos adicionales específicos del evento.
	Payload map[string]any `json:"payload,omitempty" bson:"payload,omitempty"`

	// CreatedAt marca el momento en que se creó el documento en MongoDB.
	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// ToDomain convierte el modelo de Mongo a la entidad del dominio.
func (be *BrowserEvent) ToDomain() (*domain.BrowserEvent, error) {
	if be == nil {
		return nil, errors.New("BrowserEvent model is nil")
	}

	return &domain.BrowserEvent{
		ID:            be.ID,
		EventType:     be.EventType,
		CandidateID:   be.CandidateID,
		AssessmentIDs: be.AssessmentIDs,
		Timestamp:     be.Timestamp,
		TargetID:      be.TargetID,
		Payload:       be.Payload,
	}, nil
}

// FromDomain convierte la entidad del dominio a nuestro modelo para Mongo.
func FromDomain(browserEvent *domain.BrowserEvent) (*BrowserEvent, error) {
	if browserEvent == nil {
		return nil, errors.New("browserEvent cannot be nil")
	}

	return &BrowserEvent{
		ID:            browserEvent.ID,
		EventType:     browserEvent.EventType,
		CandidateID:   browserEvent.CandidateID,
		AssessmentIDs: browserEvent.AssessmentIDs,
		Timestamp:     browserEvent.Timestamp,
		TargetID:      browserEvent.TargetID,
		Payload:       browserEvent.Payload,
		// CreatedAt se asignará externamente (por ejemplo, en el repositorio al momento de guardar).
	}, nil
}
