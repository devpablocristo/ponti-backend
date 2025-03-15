package models

import (
	"errors"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/candidate/usecases/domain"
	"gorm.io/gorm"
)

// Candidate representa el candidato en la capa de persistencia (GORM).
type Candidate struct {
	ID              string         `gorm:"primaryKey;column:id"`
	PersonID        string         `gorm:"column:person_id;not null;index"`
	Email           string         `gorm:"column:email;unique;not null"`
	ExperienceLevel string         `gorm:"column:experience_level"` // Valor textual de la experiencia (ej.: "junior")
	ExperienceRank  int            `gorm:"column:experience_rank"`  // Ranking numérico (ej.: 2)
	AssessmentsIDs  []string       `gorm:"type:text[];column:assessments_ids"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       *time.Time     `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// FromDomainCandidate convierte una entidad de dominio Candidate en el modelo Candidate.
// Realiza la conversión del slice de AssessmentIDs de domain a []string.
func FromDomainCandidate(candidate *domain.Candidate) (*Candidate, error) {
	if candidate == nil {
		return nil, errors.New("candidate is nil")
	}

	assessmentIDs := make([]string, len(candidate.AssessmentsIDs))
	for i, id := range candidate.AssessmentsIDs {
		assessmentIDs[i] = string(id)
	}

	return &Candidate{
		ID:              candidate.ID,
		PersonID:        candidate.PersonID,
		Email:           candidate.Email,
		ExperienceLevel: string(candidate.Experience.Level),
		ExperienceRank:  candidate.Experience.Rank,
		AssessmentsIDs:  assessmentIDs,
		// Los campos CreatedAt, UpdatedAt y DeletedAt se gestionan automáticamente por GORM.
	}, nil
}

// ToDomain convierte el modelo Candidate en la entidad de dominio Candidate.
// Realiza la conversión del slice de []string a []domain.AssessmentID.
func (cm *Candidate) ToDomain() (*domain.Candidate, error) {
	if cm == nil {
		return nil, errors.New("candidate model is nil")
	}

	assessmentIDs := make([]domain.AssessmentID, len(cm.AssessmentsIDs))
	for i, id := range cm.AssessmentsIDs {
		assessmentIDs[i] = domain.AssessmentID(id)
	}

	return &domain.Candidate{
		ID:       cm.ID,
		PersonID: cm.PersonID,
		Email:    cm.Email,
		Experience: domain.Experience{
			Level: domain.ExperienceLevel(cm.ExperienceLevel),
			Rank:  cm.ExperienceRank,
		},
		AssessmentsIDs: assessmentIDs,
	}, nil
}
