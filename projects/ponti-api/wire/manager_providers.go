package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
)

// ProvideManagerRepository crea la implementación concreta de manager.Repository.
func ProvideManagerRepository(repo manager.GormEnginePort) *manager.Repository {
	return manager.NewRepository(repo)
}

// ProvideManagerUseCases agrupa repositorio en manager.UseCases.
func ProvideManagerUseCases(
	rep manager.RepositoryPort,
) *manager.UseCases {
	return manager.NewUseCases(rep)
}

// ProvideManagerHandler construye el handler HTTP para Manager.
func ProvideManagerHandler(
	server manager.GinServerPort,
	UseCases manager.UseCasesPort,
	config manager.ConfigAPIPort,
	middlewares manager.MiddlewaresPort,
) *manager.Handler {
	return manager.NewHandler(UseCases, server, config, middlewares)
}

// ManagerSet expone todos los providers y bindings necesarios para Manager.
var ManagerSet = wire.NewSet(
	ProvideManagerRepository,
	ProvideManagerUseCases,
	ProvideManagerHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(manager.RepositoryPort), new(*manager.Repository)),
	wire.Bind(new(manager.UseCasesPort), new(*manager.UseCases)),
	wire.Bind(new(manager.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(manager.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(manager.MiddlewaresPort), new(*mwr.Middlewares)),
)
