package dto

import (
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

// Link es un enlace único para acceder a una evaluación.
type Link struct {
	ID           string    `json:"id"`            // Clave primaria
	AssessmentID string    `json:"assessment_id"` // Clave foránea hacia Assessment
	Token        string    `json:"token"`         // Token único para el enlace
	ExpiresAt    time.Time `json:"expires_at"`    // Fecha de expiración del enlace
	URL          string    `json:"url"`           // URL para acceder a la evaluación
}

// ToDomain convierte el DTO de Link en la entidad de dominio domain.Link
func (l Link) ToDomain() *domain.Link {
	return &domain.Link{
		ID:           l.ID,
		AssessmentID: l.AssessmentID,
		Token:        l.Token,
		ExpiresAt:    l.ExpiresAt,
		URL:          l.URL,
	}
}

// FromDomain convierte una entidad de dominio domain.Link en el DTO Link
func FromDomainToLink(link *domain.Link) Link {
	return Link{
		ID:           link.ID,
		AssessmentID: link.AssessmentID,
		Token:        link.Token,
		ExpiresAt:    link.ExpiresAt,
		URL:          link.URL,
	}
}
