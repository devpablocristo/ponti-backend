package dto

import (
	"errors"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// GetTimeline representa el DTO para el timeline que se enviará en la respuesta.
type GetTimeline struct {
	Tweet
	CreatedAt time.Time `json:"created_at"`
}

// FromDomainToGetTimeline convierte un objeto del dominio a un DTO GetTimeline.
func FromDomainToGetTimeline(tweet *domain.Tweet) (*GetTimeline, error) {
	if tweet == nil {
		return nil, errors.New("tweet cannot be nil")
	}

	// Utiliza el mapper básico para Tweet.
	baseTweet, err := FromDomain(tweet)
	if err != nil {
		return nil, err
	}

	return &GetTimeline{
		Tweet:     *baseTweet,
		CreatedAt: tweet.CreatedAt,
	}, nil
}
