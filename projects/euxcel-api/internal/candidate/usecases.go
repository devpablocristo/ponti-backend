package candidate

import (
	"context"
	"fmt"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/candidate/usecases/domain"
)

type useCases struct {
	repository Repository
}

func NewUseCases(rp Repository) UseCases {
	return &useCases{
		repository: rp,
	}
}

// CreateCandidate crea un nuevo candidato y envía un correo de notificación.
func (u *useCases) CreateCandidate(ctx context.Context, candidate *domain.Candidate) (string, error) {
	candidateID, err := u.repository.CreateCandidate(ctx, candidate)
	if err != nil {
		return "", fmt.Errorf("failed to create candidate: %w", err)
	}
	return candidateID, nil
}

func (u *useCases) ListCandidates(ctx context.Context) ([]domain.Candidate, error) {
	db, err := u.repository.ListCandidates(ctx)
	return db, err
}

func (u *useCases) GetCandidate(ctx context.Context, candidateID string) (*domain.Candidate, error) {
	candidate, err := u.repository.GetCandidate(ctx, candidateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find candidate by CandidateID: %w", err)
	}
	return candidate, nil
}

func (u *useCases) DeleteCandidate(ctx context.Context, ID string) error {
	return u.repository.DeleteCandidate(ctx, ID)
}

func (u *useCases) UpdateCandidate(ctx context.Context, updatedCandidate *domain.Candidate) error {
	return u.repository.UpdateCandidate(ctx, updatedCandidate)
}
