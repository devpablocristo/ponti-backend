// Package pkgmwr contiene middlewares HTTP para Gin.
package pkgmwr

type contextKey string

// Centraliza todas las constantes de headers HTTP, contexto y claves de entorno.
const (
	HeaderAPIKey       = "X-API-KEY"
	HeaderUserID       = "X-USER-ID"
	ContextUserID      = "userID"
	ContextUserIDKey   = contextKey("userID")
	ContextAPIKey      = "apiKey"
	EnvAPIKey          = "X_API_KEY"
	ContextCredentials = "credentials"
)
