package pkgmwr

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireUserIDHeader ensures a valid user ID header is present.
func RequireUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetHeader(HeaderUserID))
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "MISSING_USER_ID",
				"message": "User ID is required in '" + HeaderUserID + "' header.",
			})
			return
		}
		c.Set(ContextUserID, userID)
		c.Next()
	}
}
