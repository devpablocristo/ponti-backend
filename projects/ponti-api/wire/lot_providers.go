package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	crop "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop"
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
	cropUC crop.UseCasesPort,
) *lot.UseCases {
	return lot.NewUseCases(rep, cropUC)
}

// ProvideLotUseCasesPort adapta *lot.UseCases a la interfaz lot.UseCasesPort.
func ProvideLotUseCasesPort(uc *lot.UseCases) lot.UseCasesPort {
	return uc
}

// ProvideLotHandler construye el handler HTTP para Lot.
func ProvideLotHandler(
	server lot.GinServerPort,
	useCases lot.UseCasesPort,
	cfg lot.ConfigAPIPort,
	middlewares lot.MiddlewaresPort,
) *lot.Handler {
	return lot.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideLotAPIConfig extrae la configuración específica de API para Lot.
func ProvideLotAPIConfig(cfg *config.ConfigSet) lot.ConfigAPIPort {
	return &cfg.API
}

// ProvideLotGormEnginePort adapta *pgorm.Repository a lot.GormEnginePort.
func ProvideLotGormEnginePort(repo *pgorm.Repository) lot.GormEnginePort {
	return repo
}

// ProvideLotGinServerPort adapta *pgin.Server a lot.GinServerPort.
func ProvideLotGinServerPort(srv *pgin.Server) lot.GinServerPort {
	return srv
}

// ProvideLotMiddlewaresPort adapta *mwr.Middlewares a lot.MiddlewaresPort.
func ProvideLotMiddlewaresPort(m *mwr.Middlewares) lot.MiddlewaresPort {
	return m
}

// LotSet expone todos los providers necesarios para Lot.
var LotSet = wire.NewSet(
	ProvideLotRepository,
	ProvideLotRepositoryPort,
	ProvideLotUseCases,
	ProvideLotUseCasesPort,
	ProvideLotHandler,
	ProvideLotAPIConfig,
	ProvideLotGormEnginePort,
	ProvideLotGinServerPort,
	ProvideLotMiddlewaresPort,
)
