package wire

import (
	"github.com/google/wire"

	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	stock "github.com/devpablocristo/ponti-backend/internal/stock"
	supply "github.com/devpablocristo/ponti-backend/internal/supply"
)

// ProvideSupplyRepository crea la implementación concreta de supply.Repository.
func ProvideSupplyRepository(repo supply.GormEnginePort) *supply.Repository {
	return supply.NewRepository(repo)
}

// ProvideSupplyRepositoryPort adapta *supply.Repository a la interfaz supply.RepositoryPort.
func ProvideSupplyRepositoryPort(r *supply.Repository) supply.RepositoryPort {
	return r
}

// ProvideSupplyExporterPort entrega el exporter CSV.
func ProvideSupplyExporterPort() supply.ExporterAdapterPort {
	return supply.NewCSVExporter()
}

// ProvideSupplyUseCases agrupa repositorios en supply.UseCases.
func ProvideSupplyUseCases(
	repo supply.RepositoryPort,
	exp supply.ExporterAdapterPort,
	stockUseCases supply.StockUseCasesPort,
) *supply.UseCases {
	return supply.NewUseCases(repo, exp, stockUseCases)
}

// ProvideSupplyUseCasesPort adapta *supply.UseCases a la interfaz supply.UseCasesPort.
func ProvideSupplyUseCasesPort(uc *supply.UseCases) supply.UseCasesPort {
	return uc
}

// ProvideSupplyHandler construye el handler HTTP para Supply.
func ProvideSupplyHandler(
	server supply.GinEnginePort,
	useCases supply.UseCasesPort,
	cfg supply.ConfigAPIPort,
	middlewares supply.MiddlewaresEnginePort,
) *supply.Handler {
	return supply.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideSupplyConfigAPI extrae la configuración específica de API para Supply.
func ProvideSupplyConfigAPI(cfg *config.Config) supply.ConfigAPIPort {
	return &cfg.API
}

// ProvideSupplyGormEnginePort adapta *pgorm.Repository a supply.GormEnginePort.
func ProvideSupplyGormEnginePort(r *pgorm.Repository) supply.GormEnginePort {
	return r
}

// ProvideSupplyGinEnginePort adapta *pgin.Server a supply.GinEnginePort.
func ProvideSupplyGinEnginePort(s *pgin.Server) supply.GinEnginePort {
	return s
}

// ProvideSupplyMiddlewaresEnginePort adapta *mwr.Middlewares a supply.MiddlewaresEnginePort.
func ProvideSupplyMiddlewaresEnginePort(m *mwr.Middlewares) supply.MiddlewaresEnginePort {
	return m
}

func ProvideSupplyStockUseCasesPort(uc *stock.UseCases) supply.StockUseCasesPort {
	return uc
}

// SupplySet expone todos los providers necesarios para Supply.
var SupplySet = wire.NewSet(
	ProvideSupplyRepository,
	ProvideSupplyRepositoryPort,
	ProvideSupplyUseCases,
	ProvideSupplyUseCasesPort,
	ProvideSupplyHandler,
	ProvideSupplyConfigAPI,
	ProvideSupplyGormEnginePort,
	ProvideSupplyGinEnginePort,
	ProvideSupplyMiddlewaresEnginePort,
	ProvideSupplyStockUseCasesPort,
	ProvideSupplyExporterPort,
)
