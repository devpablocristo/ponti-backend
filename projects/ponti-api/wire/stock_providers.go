package wire

import (
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	"github.com/google/wire"
)

// ProvideStockRepository creates the concrete implementation of stock.Repository.
func ProvideStockRepository(repo stock.GormEnginePort) *stock.Repository {
	return stock.NewRepository(repo)
}

func ProvideStockRepositoryPort(r *stock.Repository) stock.RepositoryPort {
	return r
}

// ProvideStockUseCases groups repository into stock.UseCases.
func ProvideStockUseCases(rep stock.RepositoryPort) *stock.UseCases {
	return stock.NewUseCases(rep)
}

func ProvideStockUseCasesPort(uc *stock.UseCases) stock.UseCasesPort {
	return uc
}

func ProvideStockHandler(
	server stock.GinEnginePort,
	useCases stock.UseCasesPort,
	cfg stock.ConfigAPIPort,
	middlewares stock.MiddlewaresEnginePort,
	ucps project.UseCasesPort,
) *stock.Handler {
	return stock.NewHandler(useCases, server, cfg, middlewares, ucps)
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
)
