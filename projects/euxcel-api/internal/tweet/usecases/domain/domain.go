package domain

import (
	"errors"
	"time"
)

const MaxTweetLength = 280

type Tweet struct {
	ID        string
	UserID    string
	Content   string
	CreatedAt time.Time
}

func NewTweet(userID string, content string) (*Tweet, error) {
	if len(content) == 0 {
		return nil, errors.New("tweet content cannot be empty")
	}
	if len(content) > MaxTweetLength {
		return nil, errors.New("tweet content exceeds the character limit")
	}
	return &Tweet{
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}
