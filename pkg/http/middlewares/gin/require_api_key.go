package pkgmwr

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// RequireAPIKey asegura que el request tenga un API key válido en el header.
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader(HeaderAPIKey))
		expectedKey := os.Getenv(EnvAPIKey)
		if apiKey == "" {
			domErr := pkgtypes.NewError(pkgtypes.ErrAuthentication, "api key is required", nil)
			apiErr, status := pkgtypes.NewAPIError(domErr)
			c.AbortWithStatusJSON(status, apiErr.ToResponse())
			return
		}
		if expectedKey == "" || apiKey != expectedKey {
			domErr := pkgtypes.NewError(pkgtypes.ErrAuthentication, "api key is invalid", nil)
			apiErr, status := pkgtypes.NewAPIError(domErr)
			c.AbortWithStatusJSON(status, apiErr.ToResponse())
			return
		}
		c.Set(ContextAPIKey, apiKey)
		c.Next()
	}
}
