package wire

import (
	"github.com/google/wire"

	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	config "github.com/devpablocristo/ponti-backend/cmd/config"

	lot "github.com/devpablocristo/ponti-backend/internal/lot"
)

// ProvideLotRepository crea la implementación concreta de lot.Repository.
func ProvideLotRepository(repo lot.GormEnginePort) *lot.Repository {
	return lot.NewRepository(repo)
}

// ProvideLotRepositoryPort adapta *lot.Repository a la interfaz lot.RepositoryPort.
func ProvideLotRepositoryPort(r *lot.Repository) lot.RepositoryPort {
	return r
}

// ProvideLotExporterPort entrega el exporter CSV.
func ProvideLotExporterPort() lot.ExporterAdapterPort {
	return lot.NewCSVExporter()
}

// ProvideLotUseCases agrupa repositorio y exporter CSV en lot.UseCases.
func ProvideLotUseCases(
	rep lot.RepositoryPort,
	exp lot.ExporterAdapterPort,
) *lot.UseCases {
	return lot.NewUseCases(rep, exp)
}

// ProvideLotUseCasesPort adapta *lot.UseCases a la interfaz lot.UseCasesPort.
func ProvideLotUseCasesPort(uc *lot.UseCases) lot.UseCasesPort {
	return uc
}

// ProvideLotHandler construye el handler HTTP para Lot.
func ProvideLotHandler(
	server lot.GinEnginePort,
	useCases lot.UseCasesPort,
	cfg lot.ConfigAPIPort,
	middlewares lot.MiddlewaresEnginePort,
) *lot.Handler {
	return lot.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideLotConfigAPI extrae la configuración específica de API para Lot.
func ProvideLotConfigAPI(cfg *config.Config) lot.ConfigAPIPort {
	return &cfg.API
}

// ProvideLotGormEnginePort adapta *pgorm.Repository a lot.GormEnginePort.
func ProvideLotGormEnginePort(r *pgorm.Repository) lot.GormEnginePort {
	return r
}

// ProvideLotGinEnginePort adapta *pgin.Server a lot.GinEnginePort.
func ProvideLotGinEnginePort(s *pgin.Server) lot.GinEnginePort {
	return s
}

// ProvideLotMiddlewaresEnginePort adapta *mwr.Middlewares a lot.MiddlewaresEnginePort.
func ProvideLotMiddlewaresEnginePort(m *mwr.Middlewares) lot.MiddlewaresEnginePort {
	return m
}

// LotSet expone todos los providers necesarios para Lot.
var LotSet = wire.NewSet(
	ProvideLotRepository,
	ProvideLotRepositoryPort,
	ProvideLotUseCases,
	ProvideLotUseCasesPort,
	ProvideLotHandler,
	ProvideLotConfigAPI,
	ProvideLotGormEnginePort,
	ProvideLotGinEnginePort,
	ProvideLotMiddlewaresEnginePort,
	ProvideLotExporterPort,
)
