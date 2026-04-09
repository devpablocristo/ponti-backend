package idp

import (
	"context"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
)

type FirebaseAdmin struct {
	client *fbauth.Client
}

func NewFirebaseAdmin(app *firebase.App) (*FirebaseAdmin, error) {
	c, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &FirebaseAdmin{client: c}, nil
}

func (a *FirebaseAdmin) CreateUser(ctx context.Context, email, password string) (string, error) {
	u, err := a.client.CreateUser(ctx, (&fbauth.UserToCreate{}).Email(email).Password(password))
	if err != nil {
		return "", err
	}
	return u.UID, nil
}

func (a *FirebaseAdmin) GetUserUIDByEmail(ctx context.Context, email string) (string, error) {
	u, err := a.client.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	return u.UID, nil
}

func (a *FirebaseAdmin) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	return a.client.PasswordResetLink(ctx, email)
}
