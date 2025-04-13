package authe

import (
	"context"
	"errors"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe/usecases/domain"
)

type useCases struct {
	cache      Cache
	jwtService JwtService
	httpClient HttpClient
}

func NewUseCases(
	ch Cache,
	js JwtService,
	hc HttpClient,

) UseCases {
	return &useCases{
		cache:      ch,
		jwtService: js,
		httpClient: hc,
	}
}

func (u *useCases) JwtLogin(ctx context.Context, username, email, password string) (*domain.Token, error) {
	return nil, nil
}

func (u *useCases) GenerateLinkTokens(ctx context.Context, userID string) (*domain.Token, error) {
	if userID == "" {
		return nil, errors.New("userID is empty")
	}

	return nil, nil
}

func (u *useCases) Auth0Login(ctx context.Context, username, email, password string) (*domain.Token, error) {
	return nil, nil
}
