package pkgmwr

import (
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/platform/http/go/httperr"
	"github.com/devpablocristo/platform/observability/go"
)

// ErrorHandling maneja errores del contexto Gin y responde con JSON formateado.
// Loggea cada error que surfaceque al cliente con el logger request-scoped.
func ErrorHandling() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Written() {
			return
		}
		if len(c.Errors) > 0 {
			ginErr := c.Errors[0]
			status, apiErr := httperr.Normalize(ginErr.Err)
			observability.LoggerFromContext(c.Request.Context()).Error("http error",
				"event", "http_error",
				"error", ginErr.Err.Error(),
				"status", status,
				"code", apiErr.Code,
				"message", apiErr.Message,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"route", c.FullPath(),
			)
			c.AbortWithStatusJSON(status, apiErr)
		}
	}
}
