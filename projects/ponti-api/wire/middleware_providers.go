// File: wire/middleware_provider.go
package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"
)

// MiddlewaresPort define el contrato para los tres grupos de middleware.
type MiddlewaresPort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

// ProvideGlobalMiddlewares son middlewares que corren en todas las rutas.
func ProvideGlobalMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		mwr.ErrorHandling(),
		mwr.RequestAndResponseLogger(mwr.HttpLoggingOptions{
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

// ProvideValidationMiddlewares son middlewares que validan payloads y headers.
func ProvideValidationMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		mwr.ValidateCredentials(),
		mwr.ValidateUserIDHeader(),
		mwr.RequireAPIKey(),
	}
}

// ProvideJwtMiddleware devuelve el middleware de JWT (puede fallar si la config es errónea).
func ProvideJwtMiddleware() (gin.HandlerFunc, error) {
	return mwr.ValidateJWT(utils.NewConfigFromEnv()), nil
}

// ProvideProtectedMiddlewares empaqueta el validador de JWT en un slice.
func ProvideProtectedMiddlewares(jwtMiddleware gin.HandlerFunc) []gin.HandlerFunc {
	return []gin.HandlerFunc{jwtMiddleware}
}

// ProvideMiddlewares agrupa los tres slices en un struct único.
func ProvideMiddlewares(
	global []gin.HandlerFunc,
	validation []gin.HandlerFunc,
	protected []gin.HandlerFunc,
) *mwr.Middlewares {
	return &mwr.Middlewares{
		Global:     global,
		Validation: validation,
		Auth:       protected,
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
	ProvideMiddlewares,
	wire.Bind(new(MiddlewaresPort), new(*mwr.Middlewares)),
)
