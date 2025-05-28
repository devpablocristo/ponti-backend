package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
)

// ProvideCropRepository crea la implementación concreta de crop.Repository.
func ProvideCropRepository(repo crop.GormEnginePort) *crop.Repository {
	return crop.NewRepository(repo)
}

// ProvideCropUseCases agrupa repositorios en crop.UseCases.
func ProvideCropUseCases(
	rep crop.RepositoryPort,
) *crop.UseCases {
	return crop.NewUseCases(rep)
}

// ProvideCropHandler construye el handler HTTP para Crop.
func ProvideCropHandler(
	server crop.GinServerPort,
	UseCases crop.UseCasesPort,
	config crop.ConfigAPIPort,
	middlewares crop.MiddlewaresPort,
) *crop.Handler {
	return crop.NewHandler(UseCases, server, config, middlewares)
}

// CropSet expone todos los providers y bindings necesarios para Crop.
var CropSet = wire.NewSet(
	ProvideCropRepository,
	ProvideCropUseCases,
	ProvideCropHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(crop.RepositoryPort), new(*crop.Repository)),
	wire.Bind(new(crop.UseCasesPort), new(*crop.UseCases)),
	wire.Bind(new(crop.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(crop.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(crop.MiddlewaresPort), new(*mwr.Middlewares)),
)
