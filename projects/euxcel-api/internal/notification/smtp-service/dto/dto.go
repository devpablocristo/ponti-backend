package dto

import (
	"errors"

	smtp "github.com/alphacodinggroup/euxcel-backend/pkg/notification/smtp"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/notification/usecases/domain"
)

type Email struct {
	Address string
	Subject string
	Body    string
}

func FromDomain(email *domain.Email) (*smtp.Email, error) {
	if email == nil {
		return nil, errors.New("email cannot be nil")
	}
	return &smtp.Email{
		Address: email.Address,
		Subject: email.Subject,
		Body:    email.Body,
	}, nil
}
