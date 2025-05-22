package pkgmwr

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey != os.Getenv("X_API_KEY") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}
		userID := c.GetHeader("X-User-Id")
		if userID != "" {
			c.Set("userID", userID)
		}

		c.Next()
	}
}
