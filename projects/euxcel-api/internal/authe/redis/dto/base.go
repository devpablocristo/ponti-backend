package dto

import (
	"encoding/json"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/authe/usecases/domain"
)

// Token representa el token de acceso para Redis
type Token struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token,omitempty"` // omitempty si es opcional
	AccessExpiresAt  time.Time `json:"access_expires_at"`       // Ahora se almacena en JSON
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`      // Ahora se almacena en JSON
	IssuedAt         time.Time `json:"issued_at"`               // Ahora se almacena en JSON
	Subject          string    `json:"subject"`                 // Ahora se almacena en JSON
	TokenType        string    `json:"token_type"`              // Ej: "Bearer"
}

// ToDomain convierte la estructura desde Redis (deserialización)
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

// ToJson convierte el token a formato JSON para almacenar en Redis
func (dto *Token) ToJson() (string, error) {
	data, err := json.Marshal(dto)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJson convierte una cadena JSON a la estructura Token
func FromJson(data string) (*Token, error) {
	var token Token
	err := json.Unmarshal([]byte(data), &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// FromJSONToDomain convierte una cadena JSON en una estructura de dominio.Token
func FromJSONToDomain(data string) (*domain.Token, error) {
	var dto Token
	err := json.Unmarshal([]byte(data), &dto)
	if err != nil {
		return nil, err
	}
	return dto.ToDomain(), nil
}

// FromDomainToJSON convierte una estructura de dominio.Token en una cadena JSON para Redis
func FromDomainToJSON(domainToken *domain.Token) (string, error) {
	dto := &Token{
		AccessToken:      domainToken.AccessToken,
		RefreshToken:     domainToken.RefreshToken,
		AccessExpiresAt:  domainToken.AccessExpiresAt,
		RefreshExpiresAt: domainToken.RefreshExpiresAt,
		IssuedAt:         domainToken.IssuedAt,
		Subject:          domainToken.Subject,
		TokenType:        domainToken.TokenType,
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
