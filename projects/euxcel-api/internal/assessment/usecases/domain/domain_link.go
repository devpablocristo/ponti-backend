package domain

import "time"

// Link es un enlace único para acceder a una evaluación.
type Link struct {
	ID           string    // Clave primaria
	AssessmentID string    // Clave foránea hacia Assessment
	Token        string    // Token único para el enlace
	ExpiresAt    time.Time // Fecha de expiración del enlace
	URL          string    // URL para acceder a la evaluación
}
