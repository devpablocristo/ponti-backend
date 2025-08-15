package wire

import (
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement"
	"github.com/google/wire"
)

func ProvideSupplyMovementRepository(repo supply_movement.GormEnginePort) *supply_movement.Repository {
	return supply_movement.NewRepository(repo)
}

func ProvideSupplyMovementRepositoryPort(r *supply_movement.Repository) supply_movement.RepositoryPort {
	return r
}

func ProvideSupplyMovementUseCases(suc *stock.UseCases, r *supply_movement.Repository) *supply_movement.UseCases {
	return supply_movement.NewUseCases(r, suc)
}

func ProvideSupplyMovementUseCasesPort(uc *supply_movement.UseCases) supply_movement.UseCasesPort {
	return uc
}

func ProvideSupplyMovementHandler(
	server supply_movement.GinEnginePort,
	useCases supply_movement.UseCasesPort,
	cfg supply_movement.ConfigAPIPort,
	middlewares supply_movement.MiddlewaresEnginePort,
	ucps project.UseCasesPort,
) *supply_movement.Handler {
	return supply_movement.NewHandler(useCases, server, cfg, middlewares, ucps)
}

func ProvideSupplyMovementConfigAPI(cfg *config.Config) supply_movement.ConfigAPIPort {
	return &cfg.API
}

func ProvideSupplyMovementGormEnginePort(r *pgorm.Repository) supply_movement.GormEnginePort {
	return r
}

func ProvideSupplyMovementGinEnginePort(s *pgin.Server) supply_movement.GinEnginePort {
	return s
}

func ProvideSupplyMovementMiddlewaresEnginePort(m *mwr.Middlewares) supply_movement.MiddlewaresEnginePort {
	return m
}

var SupplyMovementSet = wire.NewSet(
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
