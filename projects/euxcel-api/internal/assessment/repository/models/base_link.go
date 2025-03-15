package models

import (
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

// Link es un enlace único para acceder a una evaluación.
type Link struct {
	ID           string    `gorm:"primaryKey"`                             // Clave primaria
	AssessmentID string    `gorm:"index;not null"`                         // Clave foránea hacia Assessment
	Token        string    `gorm:"type:varchar(255);not null;uniqueIndex"` // Token único para el enlace
	ExpiresAt    time.Time `gorm:"not null"`                               // Fecha de expiración del enlace
	URL          string    `gorm:"type:text;not null"`                     // URL para acceder a la evaluación
	CreatedAt    time.Time `gorm:"autoCreateTime"`                         // Fecha de creación
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`                         // Fecha de última actualización
}

func FromDomainToLink(domainLink *domain.Link) *Link {
	return &Link{
		ID:           domainLink.ID,
		AssessmentID: domainLink.AssessmentID,
		Token:        domainLink.Token,
		ExpiresAt:    domainLink.ExpiresAt,
		URL:          domainLink.URL,
	}
}

// ToDomain convierte el modelo Link en la entidad de dominio (domain.Link).
// Este es un método del modelo.
func (l Link) ToDomain() *domain.Link {
	return &domain.Link{
		ID:           l.ID,
		AssessmentID: l.AssessmentID,
		Token:        l.Token,
		ExpiresAt:    l.ExpiresAt,
		URL:          l.URL,
	}
}
