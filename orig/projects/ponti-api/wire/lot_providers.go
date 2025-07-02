package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
)

// ProvideLotRepository crea la implementación concreta de lot.Repository.
func ProvideLotRepository(repo lot.GormEnginePort) *lot.Repository {
	return lot.NewRepository(repo)
}

// ProvideLotRepositoryPort adapta *lot.Repository a la interfaz lot.RepositoryPort.
func ProvideLotRepositoryPort(r *lot.Repository) lot.RepositoryPort {
	return r
}

// ProvideLotUseCases agrupa repositorio y servicio de crop en lot.UseCases.
func ProvideLotUseCases(
	rep lot.RepositoryPort,
) *lot.UseCases {
	return lot.NewUseCases(rep)
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
func ProvideLotConfigAPI(cfg *config.AllConfigs) lot.ConfigAPIPort {
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
)
