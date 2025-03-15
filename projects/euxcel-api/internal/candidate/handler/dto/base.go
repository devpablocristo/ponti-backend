package dto

import (
	"errors"

	"github.com/devpablocristo/monorepo/projects/qh/internal/candidate/usecases/domain"
)

// Candidate representa el DTO para el candidato.
type Candidate struct {
	ID              string   `json:"id,omitempty"`
	PersonID        string   `json:"person_id" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	ExperienceLevel string   `json:"experience_level" validate:"required"`
	ExperienceRank  int      `json:"experience_rank" validate:"gte=0"`
	AssessmentsIDs  []string `json:"assessments_ids"`
}

// ToDomain convierte el DTO Candidate a su equivalente en dominio.
// Retorna nil si el receptor es nil.
func (c *Candidate) ToDomain() *domain.Candidate {
	if c == nil {
		return nil
	}

	assessmentIDs := make([]domain.AssessmentID, len(c.AssessmentsIDs))
	for i, id := range c.AssessmentsIDs {
		assessmentIDs[i] = domain.AssessmentID(id)
	}

	return &domain.Candidate{
		ID:       c.ID,
		PersonID: c.PersonID,
		Email:    c.Email,
		Experience: domain.Experience{
			Level: domain.ExperienceLevel(c.ExperienceLevel),
			Rank:  c.ExperienceRank,
		},
		AssessmentsIDs: assessmentIDs,
	}
}

func FromDomain(candidate *domain.Candidate) (*Candidate, error) {
	if candidate == nil {
		return nil, errors.New("candidate cannot be nil")
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
	}, nil
}
