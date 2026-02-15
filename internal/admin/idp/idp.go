package idp

import "context"

// AdminClient abstracts Identity Platform user administration.
// Domain/AuthZ does NOT depend on this interface; it is used only by admin endpoints.
type AdminClient interface {
	CreateUser(ctx context.Context, email, password string) (uid string, err error)
	GetUserUIDByEmail(ctx context.Context, email string) (uid string, err error)
	GeneratePasswordResetLink(ctx context.Context, email string) (string, error)
}

