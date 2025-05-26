package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	mdw "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

func ProvideGlobalMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		mdw.ErrorHandling(),
		mdw.RequestAndResponseLogger(mdw.HttpLoggingOptions{
			LogLevel:       "info",
			IncludeHeaders: true,
			IncludeBody:    false,
			ExcludedPaths: []string{
				"/health",
				"/ping",
				"/swagger/spec",
				"/swagger/ui/index.html",
			},
		}),
	}
}

func ProvideValidationMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		mdw.ValidateCredentials(),
		mdw.ValidateUserIDHeader(),
	}
}

func ProvideJwtMiddleware() (gin.HandlerFunc, error) {
	return mdw.ValidateJWT(
		utils.NewConfigFromEnv(),
	), nil
}

func ProvideProtectedMiddlewares(jwtMiddleware gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		jwtMiddleware,
	}
}

var GlobalMiddlewareSet = wire.NewSet(
	ProvideGlobalMiddlewares,
)

var ValidationMiddlewareSet = wire.NewSet(
	ProvideValidationMiddlewares,
)

var AuthMiddlewareSet = wire.NewSet(
	ProvideJwtMiddleware,
	ProvideProtectedMiddlewares,
)
