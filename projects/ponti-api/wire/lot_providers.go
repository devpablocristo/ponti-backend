package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
)

// ProvideLotRepository crea la implementación concreta de lot.Repository.
func ProvideLotRepository(repo lot.GormEnginePort) *lot.Repository {
	return lot.NewRepository(repo)
}

// ProvideLotUseCases agrupa repositorio y servicio de crop en lot.UseCases.
func ProvideLotUseCases(
	rep lot.RepositoryPort,
	cropUC crop.UseCasesPort,
) *lot.UseCases {
	return lot.NewUseCases(rep, cropUC)
}

// ProvideLotHandler construye el handler HTTP para Lot.
func ProvideLotHandler(
	server lot.GinServerPort,
	UseCases lot.UseCasesPort,
	config lot.ConfigAPIPort,
	middlewares lot.MiddlewaresPort,
) *lot.Handler {
	return lot.NewHandler(UseCases, server, config, middlewares)
}

// LotSet expone todos los providers y bindings necesarios para Lot.
var LotSet = wire.NewSet(
	ProvideLotRepository,
	ProvideLotUseCases,
	ProvideLotHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(lot.RepositoryPort), new(*lot.Repository)),
	wire.Bind(new(lot.UseCasesPort), new(*lot.UseCases)),
	wire.Bind(new(lot.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(lot.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(lot.MiddlewaresPort), new(*mwr.Middlewares)),
)
