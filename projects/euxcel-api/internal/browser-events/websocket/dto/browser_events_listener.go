package dto

import (
	"time"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/browser-events/usecases/domain"
)

// BrowserEventDTO es un DTO genérico para representar cualquier evento del navegador.
type BrowserEvent struct {
	// EventType indica el tipo de evento, por ejemplo: "click", "keydown", "paste", etc.
	EventType string `json:"eventType"`
	// CandidateID identifica al candidato relacionado con el evento.
	CandidateID string `json:"candidateId"`
	// AssessmentIDs contiene los IDs de las evaluaciones asociadas.
	AssessmentIDs []string `json:"assessmentIds"`
	// Timestamp marca el momento en que ocurrió el evento.
	Timestamp time.Time `json:"timestamp"`
	// TargetID es el identificador del elemento HTML sobre el que se produjo el evento (opcional).
	TargetID string `json:"targetId,omitempty"`
	// Payload contiene datos adicionales específicos del evento.
	Payload map[string]any `json:"payload,omitempty"`
}

func (b *BrowserEvent) ToDomain() *domain.BrowserEvent {
	return &domain.BrowserEvent{
		CandidateID:   b.CandidateID,
		AssessmentIDs: b.AssessmentIDs,
		EventType:     b.EventType,
		Timestamp:     b.Timestamp,
		TargetID:      b.TargetID,
		Payload:       b.Payload,
	}
}
