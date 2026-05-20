package wire

import (
	"github.com/devpablocristo/ponti-backend/cmd/config"
	"github.com/devpablocristo/ponti-backend/internal/labor"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
	"github.com/devpablocristo/ponti-backend/internal/project"
	"github.com/google/wire"
)

// ProvideLaborRepository crea la implementación concreta de labor.Repository.
func ProvideLaborRepository(repo labor.GormEnginePort) *labor.Repository {
	return labor.NewRepository(repo)
}

func ProvideLaborRepositoryPort(r *labor.Repository) labor.RepositoryPort {
	return r
}

// ProvideLaborExporterPort entrega el exporter CSV.
func ProvideLaborExporterPort() labor.ExporterAdapterPort {
	return labor.NewCSVExporter()
}

// ProvideLaborUseCases agrupa repositorio y exporter CSV.
func ProvideLaborUseCases(rep labor.RepositoryPort, exp labor.ExporterAdapterPort, projectUC project.UseCasesPort) *labor.UseCases {
	return labor.NewUseCases(rep, exp, projectUC)
}

func ProvideLaborUseCasesPort(uc *labor.UseCases) labor.UseCasesPort {
	return uc
}

func ProvideLaborHandler(
	server labor.GinEnginePort,
	useCases labor.UseCasesPort,
	cfg labor.ConfigAPIPort,
	middlewares labor.MiddlewaresEnginePort) *labor.Handler {
	return labor.NewHandler(useCases, server, cfg, middlewares)
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
	ProvideLaborExporterPort,
)
