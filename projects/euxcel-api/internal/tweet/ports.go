package tweet

import (
	"context"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// UseCases define las operaciones de la capa de negocio para tweets.
type UseCases interface {
	CreateTweet(context.Context, *domain.Tweet) (string, error)
	GetTimeline(context.Context, string) ([]domain.Tweet, error)
}

// Repository define las operaciones necesarias en Cassandra.
type Repository interface {
	SaveTweet(context.Context, *domain.Tweet) (string, error)
	GetTweetsByUserIDs(context.Context, []string, int, int) ([]domain.Tweet, error)
	InsertTweetIntoTimeline(context.Context, string, *domain.Tweet) error
}

// Cache define la interfaz para las operaciones de caché.
type Cache interface {
	InvalidateUserTimeline(context.Context, string) error
	GetTimeline(context.Context, string) ([]domain.Tweet, error)
	SetTimeline(context.Context, string, []domain.Tweet) error
	PushTweetToTimeline(context.Context, string, *domain.Tweet) error
	Close()
}

// Broker define la interfaz para la publicación de eventos (por ejemplo, en RabbitMQ).
type Broker interface {
	PublishTweetCreated(context.Context, *domain.Tweet) error
}

// Gomock
// mockgen -source=ports.go -destination=./mocks/mock_tweet.go -package=mocks
