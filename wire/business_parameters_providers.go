package wire

import (
	"github.com/google/wire"

	cfg "github.com/devpablocristo/ponti-backend/cmd/config"
	bparams "github.com/devpablocristo/ponti-backend/internal/business-parameters"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	gormpkg "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

// ProvideBusinessParametersGormEnginePort ...
func ProvideBusinessParametersGormEnginePort(r *gormpkg.Repository) bparams.GormEnginePort {
	return r
}

// ProvideBusinessParametersRepository ...
func ProvideBusinessParametersRepository(r bparams.GormEnginePort) *bparams.Repository {
	return bparams.NewRepository(r)
}

// ProvideBusinessParametersRepositoryPort ...
func ProvideBusinessParametersRepositoryPort(repo *bparams.Repository) bparams.RepositoryPort {
	return repo
}

// ProvideBusinessParametersConfigAPI ...
func ProvideBusinessParametersConfigAPI(c *cfg.Config) bparams.ConfigAPIPort {
	return &c.API
}

// ProvideBusinessParametersGinEnginePort ...
func ProvideBusinessParametersGinEnginePort(s *pgin.Server) bparams.GinEnginePort {
	return s
}

// ProvideBusinessParametersMiddlewaresEnginePort ...
func ProvideBusinessParametersMiddlewaresEnginePort(m *mwr.Middlewares) bparams.MiddlewaresEnginePort {
	return m
}

// ProvideBusinessParametersUseCases ...
func ProvideBusinessParametersUseCases(rep bparams.RepositoryPort) *bparams.UseCases {
	return bparams.NewUseCases(rep)
}

// ProvideBusinessParametersUseCasesPort ...
func ProvideBusinessParametersUseCasesPort(u *bparams.UseCases) bparams.UseCasesPort {
	return u
}

// ProvideBusinessParametersHandler ...
func ProvideBusinessParametersHandler(
	server bparams.GinEnginePort,
	ucs bparams.UseCasesPort,
	cfg bparams.ConfigAPIPort,
	mws bparams.MiddlewaresEnginePort,
) *bparams.Handler {
	return bparams.NewHandler(ucs, server, cfg, mws)
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
