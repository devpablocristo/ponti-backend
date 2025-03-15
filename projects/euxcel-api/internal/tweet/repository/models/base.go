package models

import (
	"fmt"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// Tweet representa el modelo de datos para Cassandra.
type Tweet struct {
	ID        string    // Identificador único.
	UserID    string    // Identificador del usuario creador.
	Content   string    // Contenido del tweet (máximo 280 caracteres).
	CreatedAt time.Time // Fecha y hora de creación.
}

// ToDomain convierte una instancia de models.Tweet a un domain.Tweet.
func (m *Tweet) ToDomain() (*domain.Tweet, error) {
	if m == nil {
		return nil, fmt.Errorf("ToDomain: tweet model is nil")
	}
	return &domain.Tweet{
		ID:        m.ID,
		UserID:    m.UserID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}, nil
}

// FromDomain crea una instancia de models.Tweet a partir de un domain.Tweet.
func FromDomain(d *domain.Tweet) (*Tweet, error) {
	if d == nil {
		return nil, fmt.Errorf("FromDomain: domain tweet is nil")
	}
	return &Tweet{
		ID:        d.ID,
		UserID:    d.UserID,
		Content:   d.Content,
		CreatedAt: d.CreatedAt,
	}, nil
}

// ToDomainSlice convierte un slice de models.Tweet a un slice de domain.Tweet.
func ToDomainSlice(cassTweets []Tweet) ([]domain.Tweet, error) {
	domainTweets := make([]domain.Tweet, len(cassTweets))
	for i, m := range cassTweets {
		dt, err := m.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("ToDomainSlice: error converting tweet at index %d: %w", i, err)
		}
		if dt == nil {
			return nil, fmt.Errorf("ToDomainSlice: converted tweet at index %d is nil", i)
		}
		domainTweets[i] = *dt
	}
	return domainTweets, nil
}

// FromDomainSlice convierte un slice de domain.Tweet a un slice de models.Tweet.
func FromDomainSlice(domainTweets []domain.Tweet) ([]Tweet, error) {
	cassTweets := make([]Tweet, len(domainTweets))
	for i, d := range domainTweets {
		ct, err := FromDomain(&d)
		if err != nil {
			return nil, fmt.Errorf("FromDomainSlice: error converting tweet at index %d: %w", i, err)
		}
		if ct == nil {
			return nil, fmt.Errorf("FromDomainSlice: converted tweet at index %d is nil", i)
		}
		cassTweets[i] = *ct
	}
	return cassTweets, nil
}
