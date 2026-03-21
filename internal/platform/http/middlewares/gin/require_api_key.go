package pkgmwr

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	"github.com/devpablocristo/core/saas/go/shared/httperr"
)

// RequireAPIKey asegura que el request tenga un API key válido en el header.
func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := strings.TrimSpace(c.GetHeader(HeaderAPIKey))
		expectedKey := os.Getenv(EnvAPIKey)
		if apiKey == "" {
			err := domainerr.Unauthorized("api key is required")
			status, apiErr := httperr.Normalize(err)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		if expectedKey == "" || apiKey != expectedKey {
			err := domainerr.Unauthorized("api key is invalid")
			status, apiErr := httperr.Normalize(err)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		c.Set(ContextAPIKey, apiKey)
		c.Next()
	}
}
