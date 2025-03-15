package dto

import "time"

// Response
type LoginResponse struct {
	AccessToken     string    `json:"access_token"`
	AccessExpiresAt time.Time `json:"access_expired_at"`
}
