package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
)

// ProvideCustomerRepository crea la implementación concreta de customer.Repository.
func ProvideCustomerRepository(repo customer.GormEnginePort) *customer.Repository {
	return customer.NewRepository(repo)
}

// ProvideCustomerUseCases agrupa repositorios en customer.UseCases.
func ProvideCustomerUseCases(
	rep customer.RepositoryPort,
) *customer.UseCases {
	return customer.NewUseCases(rep)
}

// ProvideCustomerHandler construye el handler HTTP para Customer.
func ProvideCustomerHandler(
	server customer.GinServerPort,
	UseCases customer.UseCasesPort,
	config customer.ConfigAPIPort,
	middlewares customer.MiddlewaresPort,
) *customer.Handler {
	return customer.NewHandler(UseCases, server, config, middlewares)
}

// CustomerSet expone todos los providers y bindings necesarios para Customer.
var CustomerSet = wire.NewSet(
	ProvideCustomerRepository,
	ProvideCustomerUseCases,
	ProvideCustomerHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(customer.RepositoryPort), new(*customer.Repository)),
	wire.Bind(new(customer.UseCasesPort), new(*customer.UseCases)),
	wire.Bind(new(customer.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(customer.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(customer.MiddlewaresPort), new(*mwr.Middlewares)),
)
