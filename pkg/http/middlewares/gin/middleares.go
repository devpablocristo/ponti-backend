package pkgmwr

import (
	"github.com/gin-gonic/gin"

	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

// Middlewares implementa el contrato agrupando tus tres etapas.
type Middlewares struct {
	global     []gin.HandlerFunc
	validation []gin.HandlerFunc
	protected  []gin.HandlerFunc
}

// NewDefaultMiddlewares construye el objeto completo.
// Recibe la configuración en lugar de leer variables de entorno directamente.
func NewDefaultMiddlewares() *Middlewares {
	cfg := utils.NewConfigFromEnv()

	// Global
	global := []gin.HandlerFunc{
		ErrorHandling(),
		RequestAndResponseLogger(HttpLoggingOptions{
			LogLevel:       "info",
			IncludeHeaders: true,
			IncludeBody:    false,
			ExcludedPaths:  []string{"/health", "/ping", "/swagger/spec", "/swagger/ui/index.html"},
		}),
	}

	// Validation
	validation := []gin.HandlerFunc{
		ValidateCredentials(),
		ValidateUserIDHeader(),
		RequireAPIKey(),
	}

	// Protected (JWT)
	jwtMw := ValidateJWT(cfg)
	protected := []gin.HandlerFunc{jwtMw}

	return &Middlewares{
		global:     global,
		validation: validation,
		protected:  protected,
	}
}

// GetGlobal devuelve los middlewares globales (logs, errores…)
func (m *Middlewares) GetGlobal() []gin.HandlerFunc {
	return m.global
}

// GetValidation devuelve los middlewares de validación (payload, headers…)
func (m *Middlewares) GetValidation() []gin.HandlerFunc {
	return m.validation
}

// GetProtected devuelve los middlewares de protección (JWT…)
func (m *Middlewares) GetProtected() []gin.HandlerFunc {
	return m.protected
}
