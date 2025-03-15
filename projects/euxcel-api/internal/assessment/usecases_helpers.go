package assessment

import (
	"context"
	"fmt"
)

func (u *useCases) buildEmail(ctx context.Context, LinkID string) (string, string, string, error) {
	// 1. Obtener el Link
	link, err := u.repository.GetLink(ctx, LinkID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get assessment link: %w", err)
	}

	// 2. Obtener la evaluación
	assessment, err := u.repository.GetAssessment(ctx, link.AssessmentID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get assessment: %w", err)
	}

	// 3. Obtener los datos del candidato
	candidate, err := u.candidateUc.GetCandidate(ctx, assessment.CandidateID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get candidate: %w", err)
	}

	// 4. Obtener la información personal del candidato
	candidatePersonalInfo, err := u.personUc.GetPerson(ctx, candidate.PersonID)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get person: %w", err)
	}

	// 5. Construir el correo
	assessmentCfg := u.config.GetAssessmentConfig()

	address := candidate.Email
	subject := assessmentCfg.Subject
	bodyTemplate := assessmentCfg.BodyTemplate
	linkURL := fmt.Sprintf("%s?token=%s", assessmentCfg.BaseURL, link.Token)
	body := fmt.Sprintf(
		"Hola %s,\n\n%s",
		candidatePersonalInfo.FirstName,
		fmt.Sprintf(bodyTemplate, linkURL),
	)

	return address, subject, body, nil
}
