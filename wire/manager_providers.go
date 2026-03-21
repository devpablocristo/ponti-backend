package wire

import (
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	manager "github.com/devpablocristo/ponti-backend/internal/manager"
)

// ProvideManagerRepository crea la implementación concreta de manager.Repository.
func ProvideManagerRepository(repo manager.GormEnginePort) *manager.Repository {
	return manager.NewRepository(repo)
}

// ProvideManagerRepositoryPort adapta *manager.Repository a la interfaz manager.RepositoryPort.
func ProvideManagerRepositoryPort(r *manager.Repository) manager.RepositoryPort {
	return r
}

// ProvideManagerUseCases agrupa repositorio en manager.UseCases.
func ProvideManagerUseCases(
	rep manager.RepositoryPort,
) *manager.UseCases {
	return manager.NewUseCases(rep)
}

// ProvideManagerUseCasesPort adapta *manager.UseCases a la interfaz manager.UseCasesPort.
func ProvideManagerUseCasesPort(uc *manager.UseCases) manager.UseCasesPort {
	return uc
}

// ProvideManagerHandler construye el handler HTTP para Manager.
func ProvideManagerHandler(
	server manager.GinEnginePort,
	useCases manager.UseCasesPort,
	cfg manager.ConfigAPIPort,
	middlewares manager.MiddlewaresEnginePort,
) *manager.Handler {
	return manager.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideManagerConfigAPI extrae la configuración específica de API para Manager.
func ProvideManagerConfigAPI(cfg *config.Config) manager.ConfigAPIPort {
	return &cfg.API
}

// ProvideManagerGormEnginePort adapta *pgorm.Repository a manager.GormEnginePort.
func ProvideManagerGormEnginePort(r *pgorm.Repository) manager.GormEnginePort {
	return r
}

// ProvideManagerGinEnginePort adapta *pgin.Server a manager.GinEnginePort.
func ProvideManagerGinEnginePort(s *pgin.Server) manager.GinEnginePort {
	return s
}

// ProvideManagerMiddlewaresEnginePort adapta *mwr.Middlewares a manager.MiddlewaresEnginePort.
func ProvideManagerMiddlewaresEnginePort(m *mwr.Middlewares) manager.MiddlewaresEnginePort {
	return m
}

// ManagerSet expone todos los providers necesarios para Manager.
var ManagerSet = wire.NewSet(
	ProvideManagerRepository,
	ProvideManagerRepositoryPort,
	ProvideManagerUseCases,
	ProvideManagerUseCasesPort,
	ProvideManagerHandler,
	ProvideManagerConfigAPI,
	ProvideManagerGormEnginePort,
	ProvideManagerGinEnginePort,
	ProvideManagerMiddlewaresEnginePort,
)
