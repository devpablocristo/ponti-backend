package pkgmwr

import (
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/http/go/httperr"
)

// ErrorHandling maneja errores del contexto Gin y responde con JSON formateado.
func ErrorHandling() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Written() {
			return
		}
		if len(c.Errors) > 0 {
			ginErr := c.Errors[0]
			status, apiErr := httperr.Normalize(ginErr.Err)
			c.AbortWithStatusJSON(status, apiErr)
		}
	}
}
