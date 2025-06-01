package pkgmwr

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// ErrorHandling handles Gin context errors and responds with formatted JSON.
func ErrorHandling() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Written() {
			return
		}
		if len(c.Errors) > 0 {
			ginErr := c.Errors[0]
			var status int
			var response any
			var domainErr *pkgtypes.Error
			if errors.As(ginErr.Err, &domainErr) {
				apiErr, code := pkgtypes.NewAPIError(domainErr)
				response = apiErr.ToResponse()
				status = code
			} else {
				var apiErr *pkgtypes.APIError
				if errors.As(ginErr.Err, &apiErr) {
					response = apiErr.ToResponse()
					status = apiErr.Code
				} else {
					response = gin.H{
						"error":   "INTERNAL_ERROR",
						"message": "An internal error occurred. Please try again later.",
						"details": ginErr.Err.Error(),
					}
					status = http.StatusInternalServerError
				}
			}
			c.AbortWithStatusJSON(status, response)
		}
	}
}
