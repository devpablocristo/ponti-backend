// File: wire/middleware_provider.go
package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
)

type MiddlewaresPort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

func ProvideMiddlewares() *mwr.Middlewares {
	return mwr.NewDefaultMiddlewares()
}

// ProvideMiddlewaresPort convierte el *mwr.Middlewares en la interfaz MiddlewaresPort.
func ProvideMiddlewaresPort(m *mwr.Middlewares) MiddlewaresPort {
	return m
}

// MiddlewareSet expone los dos providers necesarios.
var MiddlewareSet = wire.NewSet(
	ProvideMiddlewares,
	ProvideMiddlewaresPort,
)
