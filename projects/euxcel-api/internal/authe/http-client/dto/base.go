package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/authe/usecases/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token,omitempty"` // omitempty si es opcional
	AccessExpiresAt  time.Time `json:"-"`                       // No viene en el JSON, se calcula desde el JWT
	RefreshExpiresAt time.Time `json:"-"`                       // No viene en el JSON, se calcula desde el refresh token
	IssuedAt         time.Time `json:"-"`                       // No viene en el JSON, se obtiene del claim "iat" del JWT
	Subject          string    `json:"-"`                       // No viene en el JSON, se obtiene del claim "sub" del JWT
	TokenType        string    `json:"token_type"`              // Ej: "Bearer"
}

// Método ToDomain convierte el DTO a la estructura del dominio
func (dto *Token) ToDomain() *domain.Token {
	return &domain.Token{
		AccessToken:      dto.AccessToken,
		RefreshToken:     dto.RefreshToken,
		AccessExpiresAt:  dto.AccessExpiresAt,
		RefreshExpiresAt: dto.RefreshExpiresAt,
		IssuedAt:         dto.IssuedAt,
		Subject:          dto.Subject,
		TokenType:        dto.TokenType,
	}
}

func (dto *Token) ParseJwtClaims() error {
	// Parsear access token
	token, _, err := jwt.NewParser().ParseUnverified(dto.AccessToken, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("error parsing access token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid access token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("exp claim not found or invalid")
	}
	dto.AccessExpiresAt = time.Unix(int64(exp), 0)

	if iat, ok := claims["iat"].(float64); ok {
		dto.IssuedAt = time.Unix(int64(iat), 0)
	}

	// Extraer 'sub' (opcional)
	if sub, ok := claims["sub"].(string); ok {
		dto.Subject = sub
	}

	return nil
}
