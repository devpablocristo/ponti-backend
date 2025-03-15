package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// Tweet representa el modelo que se almacenará en Redis.
type Tweet struct {
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// FromDomain convierte un objeto del dominio a un modelo para Redis.
func FromDomain(tweet *domain.Tweet) (*Tweet, error) {
	if tweet == nil {
		return nil, fmt.Errorf("tweet cannot be nil")
	}
	return &Tweet{
		UserID:    tweet.UserID,
		Content:   tweet.Content,
		CreatedAt: tweet.CreatedAt,
	}, nil
}

// ToDomain convierte un objeto del modelo de Redis a un objeto del dominio.
func (m *Tweet) ToDomain() (*domain.Tweet, error) {
	if m == nil {
		return nil, fmt.Errorf("ToDomain: tweet model is nil")
	}
	return &domain.Tweet{
		UserID:    m.UserID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}, nil
}

// ToDomainSlice convierte un slice de modelos de Redis a un slice de tweets del dominio.
func ToDomainSlice(cacheTweets []Tweet) ([]domain.Tweet, error) {
	domainTweets := make([]domain.Tweet, len(cacheTweets))
	for i, ct := range cacheTweets {
		dt, err := ct.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("ToDomainSlice: error converting tweet at index %d: %w", i, err)
		}
		domainTweets[i] = *dt
	}
	return domainTweets, nil
}

// FromDomainSlice convierte un slice de tweets del dominio a un slice de modelos de Redis.
func FromDomainSlice(domainTweets []domain.Tweet) ([]Tweet, error) {
	cacheTweets := make([]Tweet, len(domainTweets))
	for i, dt := range domainTweets {
		ct, err := FromDomain(&dt)
		if err != nil {
			return nil, fmt.Errorf("FromDomainSlice: error converting tweet at index %d: %w", i, err)
		}
		cacheTweets[i] = *ct
	}
	return cacheTweets, nil
}

// Opcional: Funciones para serialización/deserialización a JSON
func (m *Tweet) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func FromJSON(data []byte) (*Tweet, error) {
	var model Tweet
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("FromJSON: %w", err)
	}
	return &model, nil
}
