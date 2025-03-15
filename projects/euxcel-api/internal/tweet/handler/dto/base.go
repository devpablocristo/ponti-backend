package dto

import (
	"errors"

	"github.com/devpablocristo/monorepo/projects/qh/internal/tweet/usecases/domain"
)

// Tweet representa el DTO básico para un tweet.
type Tweet struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

func (t *Tweet) ToDomain() *domain.Tweet {
	return &domain.Tweet{
		UserID:  t.UserID,
		Content: t.Content,
	}
}

func FromDomain(t *domain.Tweet) (*Tweet, error) {
	if t == nil {
		return nil, errors.New("tweet cannot be nil")
	}
	return &Tweet{
		UserID:  t.UserID,
		Content: t.Content,
	}, nil
}
