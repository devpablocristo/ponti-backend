package pkgmwr

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/saas-core/shared/ctxkeys"

	pkgtypes "github.com/devpablocristo/ponti-backend/pkg/types"
)

// RequireUserIDHeader asegura que un header de ID de usuario valido este presente.
func RequireUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := strings.TrimSpace(c.GetHeader(HeaderUserID))
		if userID == "" {
			domErr := pkgtypes.NewError(pkgtypes.ErrMissingField, "user_id header is required", nil)
			apiErr, status := pkgtypes.NewAPIError(domErr)
			c.AbortWithStatusJSON(status, apiErr.ToResponse())
			return
		}
		c.Set(string(ctxkeys.Actor), userID)
		// Propagar tambien al request.Context() para capas que usan ctx estandar
		ctx := context.WithValue(c.Request.Context(), ctxkeys.Actor, userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
