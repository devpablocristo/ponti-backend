// Package pkgmwr contiene middlewares HTTP para Gin.
package pkgmwr

// Centraliza todas las constantes de headers HTTP, contexto y claves de entorno.
const (
	HeaderAPIKey = "X-API-KEY"
	HeaderUserID = "X-USER-ID"

	ContextAPIKey      = "apiKey"
	EnvAPIKey          = "X_API_KEY"
	ContextCredentials = "credentials"
)
