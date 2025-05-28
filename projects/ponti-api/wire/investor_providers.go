package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
)

// ProvideInvestorRepository crea la implementación concreta de investor.Repository.
func ProvideInvestorRepository(repo investor.GormEnginePort) *investor.Repository {
	return investor.NewRepository(repo)
}

// ProvideInvestorUseCases agrupa repositorio en investor.UseCases.
func ProvideInvestorUseCases(
	rep investor.RepositoryPort,
) *investor.UseCases {
	return investor.NewUseCases(rep)
}

// ProvideInvestorHandler construye el handler HTTP para Investor.
func ProvideInvestorHandler(
	server investor.GinServerPort,
	UseCases investor.UseCasesPort,
	config investor.ConfigAPIPort,
	middlewares investor.MiddlewaresPort,
) *investor.Handler {
	return investor.NewHandler(UseCases, server, config, middlewares)
}

// InvestorSet expone todos los providers y bindings necesarios para Investor.
var InvestorSet = wire.NewSet(
	ProvideInvestorRepository,
	ProvideInvestorUseCases,
	ProvideInvestorHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(investor.RepositoryPort), new(*investor.Repository)),
	wire.Bind(new(investor.UseCasesPort), new(*investor.UseCases)),
	wire.Bind(new(investor.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(investor.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(investor.MiddlewaresPort), new(*mwr.Middlewares)),
)
