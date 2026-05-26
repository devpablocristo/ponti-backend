package wire

import (
	"github.com/devpablocristo/ponti-backend/cmd/config"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
	"github.com/devpablocristo/ponti-backend/internal/project"
	"github.com/devpablocristo/ponti-backend/internal/stock"

	"github.com/google/wire"
)

// ProvideStockRepository crea la implementación concreta de stock.Repository.
func ProvideStockRepository(repo stock.GormEnginePort) *stock.Repository {
	return stock.NewRepository(repo)
}

func ProvideStockRepositoryPort(r *stock.Repository) stock.RepositoryPort {
	return r
}

// ProvideStockExporterPort entrega el exporter CSV.
func ProvideStockExporterPort() stock.ExporterAdapterPort {
	return stock.NewCSVExporter()
}

// ProvideStockUseCases agrupa repositorio y exporter CSV.
func ProvideStockUseCases(rep stock.RepositoryPort, exp stock.ExporterAdapterPort, projectUC project.UseCasesPort) *stock.UseCases {
	return stock.NewUseCases(rep, exp, projectUC)
}

func ProvideStockUseCasesPort(uc *stock.UseCases) stock.UseCasesPort {
	return uc
}

func ProvideStockHandler(
	server stock.GinEnginePort,
	useCases stock.UseCasesPort,
	cfg stock.ConfigAPIPort,
	middlewares stock.MiddlewaresEnginePort,
) *stock.Handler {
	return stock.NewHandler(useCases, server, cfg, middlewares)
}

func ProvideStockConfigAPI(cfg *config.Config) stock.ConfigAPIPort {
	return &cfg.API
}

func ProvideStockGormEnginePort(r *pgorm.Repository) stock.GormEnginePort {
	return r
}

func ProvideStockGinEnginePort(s *pgin.Server) stock.GinEnginePort {
	return s
}

func ProvideStockMiddlewaresEnginePort(m *mwr.Middlewares) stock.MiddlewaresEnginePort {
	return m
}

var StockSet = wire.NewSet(
	ProvideStockRepository,
	ProvideStockRepositoryPort,
	ProvideStockUseCases,
	ProvideStockUseCasesPort,
	ProvideStockHandler,
	ProvideStockConfigAPI,
	ProvideStockGormEnginePort,
	ProvideStockGinEnginePort,
	ProvideStockMiddlewaresEnginePort,
	ProvideStockExporterPort,
)
