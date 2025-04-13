package authe

import (
	"context"
	"time"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/authe/usecases/domain"
)

type UseCases interface {
	JwtLogin(context.Context, string, string, string) (*domain.Token, error)
	Auth0Login(context.Context, string, string, string) (*domain.Token, error)
	GenerateLinkTokens(context.Context, string) (*domain.Token, error)
}

type JwtService interface {
	GenerateLinkTokens(context.Context, string) (*domain.Token, error)
	ValidateToken(context.Context, string) (*domain.TokenClaims, error)
	GetAccessExpiration(context.Context) time.Duration
	GetRefreshExpiration(context.Context) time.Duration
	ExtractClaimsFromExternalToken(string) (map[string]any, error)
}

type Cache interface {
	StoreToken(context.Context, string, *domain.Token) error
	RetrieveToken(context.Context, string) (*domain.Token, error)
	Close()
}

type HttpClient interface {
	GetAccessToken(context.Context, string, any) (*domain.Token, error)
}
