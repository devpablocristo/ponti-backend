package candidate

import (
	"context"
	"time"

	domain "github.com/devpablocristo/monorepo/projects/qh/internal/candidate/usecases/domain"
)

type UseCases interface {
	CreateCandidate(context.Context, *domain.Candidate) (string, error)
	GetCandidate(context.Context, string) (*domain.Candidate, error)
	DeleteCandidate(context.Context, string) error
	ListCandidates(context.Context) ([]domain.Candidate, error)
	UpdateCandidate(context.Context, *domain.Candidate) error
}

type Repository interface {
	CreateCandidate(context.Context, *domain.Candidate) (string, error)
	UpdateCandidate(context.Context, *domain.Candidate) error
	GetCandidate(context.Context, string) (*domain.Candidate, error)
	DeleteCandidate(context.Context, string) error
	ListCandidates(context.Context) ([]domain.Candidate, error)
}

type Cache interface {
	StoreRefreshToken(context.Context, string, string, time.Time) error
	RetrieveRefreshToken(context.Context, string) (string, error)
	Close()
}

// type GrpcClient interface {
// 	Close() error
// 	RequestVerificationToken(context.Context, string) (string, string, error)
// }

// type Cache interface {
// 	StoreToken(context.Context, string, string, time.Duration) error
// 	RetrieveToken(context.Context, string) (string, error)
// 	Close()
// }

// type JwtService interface {
// 	GenerateTokens(context.Context, string) (string, string, error)
// 	ValidateToken(context.Context, string) (*sdkjwtdefs.TokenClaims, error)
// 	GetAccessExpiration() time.Duration
// 	GetRefreshExpiration() time.Duration
// }

// type MessageQueu interface {
// 	SendVerificationEmail(context.Context, string, string) error
// }
