package model

import (
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/patient/usecases/domain"
)

type CreatePatient struct {
	ID            string    `gorm:"primaryKey;type:uuid" json:"id"`
	PersonID      uint      `gorm:"not null" json:"person_id"`
	History       string    `gorm:"type:text" json:"history"`
	DiagnosisDate time.Time `gorm:"not null" json:"diagnosis_date"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// FromDomain transfiere los datos del dominio al modelo de persistencia.
func (m *CreatePatient) FromDomain(p *domain.Patient) {
	m.ID = p.ID
	m.PersonID = p.PersonID
	m.History = p.History
	m.DiagnosisDate = p.DiagnosisDate
	m.CreatedAt = p.CreatedAt
	m.UpdatedAt = p.UpdatedAt
}

// ToDomain convierte el modelo de persistencia a una entidad del dominio.
func (m *CreatePatient) ToDomain() (*domain.Patient, error) {
	return &domain.Patient{
		ID:            m.ID,
		PersonID:      m.PersonID,
		History:       m.History,
		DiagnosisDate: m.DiagnosisDate,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}, nil
}
