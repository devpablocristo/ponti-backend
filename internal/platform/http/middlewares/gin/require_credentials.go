package pkgmwr

import (
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/devpablocristo/core/http/go/httperr"

	pkgtypes "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

// RequireCredentials valida el payload de login para la autenticación del usuario.
func RequireCredentials() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var credentials pkgtypes.LoginCredentials

		if err := ctx.ShouldBindJSON(&credentials); err != nil {
			domErr := domainerr.Validation("invalid request payload")
			status, apiErr := httperr.Normalize(domErr)
			ctx.AbortWithStatusJSON(status, apiErr)
			return
		}
		if credentials.Username == "" && credentials.Email == "" {
			domErr := domainerr.Validation("The field 'username/email' is required")
			status, apiErr := httperr.Normalize(domErr)
			ctx.AbortWithStatusJSON(status, apiErr)
			return
		}

		ctx.Set(ContextCredentials, credentials)
		ctx.Next()
	}
}
