package pkgmwr

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/backend/go/contextkeys"
	"github.com/devpablocristo/core/backend/go/domainerr"
	"github.com/devpablocristo/core/backend/go/httperr"
)

// RequireUserIDHeader asegura que un header de ID de usuario valido este presente.
func RequireUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetHeader(HeaderUserID))
		if userID == "" {
			err := domainerr.Validation("user_id header is required")
			status, apiErr := httperr.Normalize(err)
			c.AbortWithStatusJSON(status, apiErr)
			return
		}
		c.Set(string(ctxkeys.Actor), userID)
		// Propagar tambien al request.Context() para capas que usan ctx estandar
		ctx := context.WithValue(c.Request.Context(), ctxkeys.Actor, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
