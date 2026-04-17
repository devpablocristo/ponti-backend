package pkgmwr

import (
	"github.com/gin-gonic/gin"

	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// RequireCredentials valida el payload de login para la autenticación del usuario.
func RequireCredentials() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var credentials pkgtypes.LoginCredentials

		if err := ctx.ShouldBindJSON(&credentials); err != nil {
			domErr := pkgtypes.NewError(pkgtypes.ErrBadRequest, "invalid request payload", err)
			apiErr, status := pkgtypes.NewAPIError(domErr)
			ctx.AbortWithStatusJSON(status, apiErr.ToResponse())
			return
		}
		if credentials.Username == "" && credentials.Email == "" {
			domErr := pkgtypes.NewMissingFieldError("username/email")
			apiErr, status := pkgtypes.NewAPIError(domErr)
			ctx.AbortWithStatusJSON(status, apiErr.ToResponse())
			return
		}

		ctx.Set(ContextCredentials, credentials)
		ctx.Next()
	}
}
