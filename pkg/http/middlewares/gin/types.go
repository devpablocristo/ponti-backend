// Package pkgmwr contiene middlewares HTTP para Gin.
package pkgmwr

type contextKey string

// Centraliza todas las constantes de headers HTTP, contexto y claves de entorno.
const (
	HeaderAPIKey       = "X-API-KEY"
	HeaderUserID       = "X-USER-ID"
	HeaderTenantID     = "X-Tenant-Id"
	ContextUserID      = "userID"
	ContextUserIDKey   = contextKey("userID")
	ContextUserEmail   = "userEmail"
	ContextUserEmailKey = contextKey("userEmail")
	ContextTenantID    = "tenantID"
	ContextTenantIDKey = contextKey("tenantID")
	ContextRoles       = "roles"
	ContextRolesKey    = contextKey("roles")
	ContextAPIKey      = "apiKey"
	EnvAPIKey          = "X_API_KEY"
	ContextCredentials = "credentials"
)
