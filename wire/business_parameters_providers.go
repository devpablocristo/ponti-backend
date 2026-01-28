package wire

import (
	"github.com/google/wire"

	cfg "github.com/alphacodinggroup/ponti-backend/cmd/config"
	business_parameters "github.com/alphacodinggroup/ponti-backend/internal/business-parameters"
	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
)

// ProvideBusinessParametersGormEnginePort ...
func ProvideBusinessParametersGormEnginePort(r *gormpkg.Repository) business_parameters.GormEnginePort {
	return r
}

// ProvideBusinessParametersRepository ...
func ProvideBusinessParametersRepository(r business_parameters.GormEnginePort) *business_parameters.Repository {
	return business_parameters.NewRepository(r)
}

// ProvideBusinessParametersRepositoryPort ...
func ProvideBusinessParametersRepositoryPort(repo *business_parameters.Repository) business_parameters.RepositoryPort {
	return repo
}

// ProvideBusinessParametersConfigAPI ...
func ProvideBusinessParametersConfigAPI(c *cfg.Config) business_parameters.ConfigAPIPort {
	return &c.API
}

// ProvideBusinessParametersGinEnginePort ...
func ProvideBusinessParametersGinEnginePort(s *pgin.Server) business_parameters.GinEnginePort {
	return s
}

// ProvideBusinessParametersMiddlewaresEnginePort ...
func ProvideBusinessParametersMiddlewaresEnginePort(m *mwr.Middlewares) business_parameters.MiddlewaresEnginePort {
	return m
}

// ProvideBusinessParametersUseCases ...
func ProvideBusinessParametersUseCases(rep business_parameters.RepositoryPort) *business_parameters.UseCases {
	return business_parameters.NewUseCases(rep)
}

// ProvideBusinessParametersUseCasesPort ...
func ProvideBusinessParametersUseCasesPort(u *business_parameters.UseCases) business_parameters.UseCasesPort {
	return u
}

// ProvideBusinessParametersHandler ...
func ProvideBusinessParametersHandler(
	server business_parameters.GinEnginePort,
	ucs business_parameters.UseCasesPort,
	cfg business_parameters.ConfigAPIPort,
	mws business_parameters.MiddlewaresEnginePort,
) *business_parameters.Handler {
	return business_parameters.NewHandler(ucs, server, cfg, mws)
}

// BusinessParametersSet ...
var BusinessParametersSet = wire.NewSet(
	ProvideBusinessParametersGormEnginePort,
	ProvideBusinessParametersRepository,
	ProvideBusinessParametersRepositoryPort,
	ProvideBusinessParametersConfigAPI,
	ProvideBusinessParametersGinEnginePort,
	ProvideBusinessParametersMiddlewaresEnginePort,

	ProvideBusinessParametersUseCases,
	ProvideBusinessParametersUseCasesPort,
	ProvideBusinessParametersHandler,
)
