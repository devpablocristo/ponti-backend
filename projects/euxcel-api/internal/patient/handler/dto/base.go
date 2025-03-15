package dto

import (
	"errors"
	"fmt"
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/patient/usecases/domain"
)

type Patient struct {
	PersonID      uint   `json:"person_id" binding:"required"`
	History       string `json:"history" binding:"required"`
	DiagnosisDate string `json:"diagnosis_date" binding:"required"`
}

func (dto *Patient) ToDomain() (*domain.Patient, error) {
	diagnosisDate, err := time.Parse("2006-01-02", dto.DiagnosisDate)
	if err != nil {
		return nil, fmt.Errorf("formato de diagnosis_date inválido: %w", err)
	}

	now := time.Now()

	patient := &domain.Patient{
		PersonID:      dto.PersonID,
		History:       dto.History,
		DiagnosisDate: diagnosisDate,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return patient, nil
}

func FromDomain(patient *domain.Patient) (*Patient, error) {
	if patient == nil {
		return nil, errors.New("patient cannot be nil")
	}
	return &Patient{
		PersonID:      patient.PersonID,
		History:       patient.History,
		DiagnosisDate: patient.DiagnosisDate.Format("2006-01-02"),
	}, nil
}
