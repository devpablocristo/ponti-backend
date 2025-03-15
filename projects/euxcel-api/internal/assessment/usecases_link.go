package assessment

import (
	"context"
	"fmt"
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

func (u *useCases) GenerateLink(ctx context.Context, assessmentID string) (string, error) {
	assessment, err := u.repository.GetAssessment(ctx, assessmentID)
	if err != nil {
		return "", fmt.Errorf("failed to get assessment by ID %s: %w", assessmentID, err)
	}

	candidate, err := u.candidateUc.GetCandidate(ctx, assessment.CandidateID)
	if err != nil {
		return "", fmt.Errorf("failed to get candidate: %w", err)
	}

	assessmentCfg := u.config.GetAssessmentConfig()
	token, err := u.autheUc.GenerateLinkTokens(ctx, candidate.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	linkURL := fmt.Sprintf("%s?token=%s", assessmentCfg.BaseURL, token.AccessToken)

	link := &domain.Link{
		AssessmentID: assessmentID,
		Token:        token.AccessToken,
		ExpiresAt:    time.Now().Add(assessmentCfg.AccessExpirationMinutes),
		URL:          linkURL,
	}

	linkID, err := u.repository.StoreLink(ctx, link)
	if err != nil {
		return "", fmt.Errorf("failed to store assessment link: %w", err)
	}

	return linkID, nil
}

func (u *useCases) SendLink(ctx context.Context, linkID string) error {
	address, subject, body, err := u.buildEmail(ctx, linkID)
	if err != nil {
		return fmt.Errorf("failed to get candidate personal info: %w", err)
	}

	if err := u.notificationUc.SendEmail(ctx, address, subject, body); err != nil {
		return fmt.Errorf("failed to send assessment link: %w", err)
	}

	return nil
}

func (u *useCases) GetLink(ctx context.Context, ID string) (*domain.Link, error) {
	link, err := u.repository.GetLink(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assessment link: %w", err)
	}
	return link, nil
}

func (u *useCases) ValidateLink(ctx context.Context, token string) (*domain.Link, error) {
	link, err := u.repository.GetLink(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get assessment link by token: %w", err)
	}

	// 2. Verificar si el link ha expirado
	if time.Now().After(link.ExpiresAt) {
		return nil, fmt.Errorf("assessment link has expired")
	}

	return link, nil
}
