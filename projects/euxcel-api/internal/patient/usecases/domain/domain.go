package domain

import "time"

// Patient representa la información clínica de un paciente en el dominio.
type Patient struct {
	ID            string    // Identificador único del paciente dentro del dominio.
	PersonID      uint      // Identificador de la persona asociada.
	History       string    // Historial clínico.
	DiagnosisDate time.Time // Fecha de diagnóstico.
	CreatedAt     time.Time // Fecha de creación en el sistema.
	UpdatedAt     time.Time // Fecha de la última actualización.
}
