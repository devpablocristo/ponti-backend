package wire

import (
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/google/wire"
)

// ProvideLaborRepository crea la implementación concreta de labor.Repository.
func ProvideLaborRepository(repo labor.GormEnginePort) *labor.Repository {
	return labor.NewRepository(repo)
}

func ProvideLaborRepositoryPort(r *labor.Repository) labor.RepositoryPort {
	return r
}

// ProvideLotUseCases agrupa repositorio y servicio de crop en lot.UseCases.
func ProvideLaborUseCases(rep labor.RepositoryPort) *labor.UseCases {
	return labor.NewUseCases(rep)
}

func ProvideLaborUseCasesPort(uc *labor.UseCases) labor.UseCasesPort {
	return uc
}

func ProvideLaborHandler(
	server labor.GinEnginePort,
	useCases labor.UseCasesPort,
	cfg labor.ConfigAPIPort,
	middlewares labor.MiddlewaresEnginePort,
	useCaseProject project.UseCasesPort) *labor.Handler {
	return labor.NewHandler(useCases, server, cfg, middlewares, useCaseProject)
}

func ProvideLaborConfigAPI(cfg *config.Config) labor.ConfigAPIPort {
	return &cfg.API
}

func ProvideLaborGormEnginePort(r *pgorm.Repository) labor.GormEnginePort {
	return r
}

func ProvideLaborGinEnginePort(s *pgin.Server) labor.GinEnginePort {
	return s
}

func ProvideLaborMiddlewaresEnginePort(m *mwr.Middlewares) labor.MiddlewaresEnginePort {
	return m
}

var LaborSet = wire.NewSet(
	ProvideLaborRepository,
	ProvideLaborRepositoryPort,
	ProvideLaborUseCases,
	ProvideLaborUseCasesPort,
	ProvideLaborHandler,
	ProvideLaborConfigAPI,
	ProvideLaborGormEnginePort,
	ProvideLaborGinEnginePort,
	ProvideLaborMiddlewaresEnginePort,
)
