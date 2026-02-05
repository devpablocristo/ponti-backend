package pkgmwr

import (
	"errors"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
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
			var apiErr *pkgtypes.APIError
			if errors.As(ginErr.Err, &apiErr) {
				c.AbortWithStatusJSON(apiErr.Code, apiErr.ToResponse())
				return
			}

			apiErr, status := pkgtypes.NewAPIError(ginErr.Err)
			c.AbortWithStatusJSON(status, apiErr.ToResponse())
		}
	}
}
