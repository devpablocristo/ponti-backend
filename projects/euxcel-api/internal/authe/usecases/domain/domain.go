package domain

import (
	"time"
)

type Token struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	IssuedAt         time.Time
	Subject          string
	TokenType        string
}

type TokenClaims struct {
	Subject   string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// type Session struct {
// 	UserUUID  string
// 	Token     Token
// 	LoggedAt  time.Time
// 	ExpiresAt time.Time
// }

// type Auth struct {
// 	UserUUID string
// 	Session  Session
// }
