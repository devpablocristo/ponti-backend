package pkgmwr

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// RequireUserIDHeader asegura que un header de ID de usuario válido esté presente.
func RequireUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetHeader(HeaderUserID))
		if userID == "" {
			domErr := pkgtypes.NewError(pkgtypes.ErrMissingField, "user_id header is required", nil)
			apiErr, status := pkgtypes.NewAPIError(domErr)
			c.AbortWithStatusJSON(status, apiErr.ToResponse())
			return
		}
		c.Set(ContextUserID, userID)
		// Propagar también al request.Context() para capas que usan ctx estándar
		ctx := context.WithValue(c.Request.Context(), ContextUserIDKey, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
