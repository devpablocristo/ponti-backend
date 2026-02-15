package idp

import (
	"context"
	"errors"
)

// NoopAdmin is used in local mode when we don't want to require ADC/GCP credentials.
type NoopAdmin struct{}

var errNoop = errors.New("identity admin disabled (local mode)")

func (n *NoopAdmin) CreateUser(ctx context.Context, email, password string) (string, error) {
	return "", errNoop
}

func (n *NoopAdmin) GetUserUIDByEmail(ctx context.Context, email string) (string, error) {
	return "", errNoop
}

func (n *NoopAdmin) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	return "", errNoop
}

