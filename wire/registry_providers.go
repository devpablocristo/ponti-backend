package wire

import (
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	registry "github.com/devpablocristo/ponti-backend/internal/registry"
)

func ProvideRegistryRepository(repo registry.GormEnginePort) *registry.Repository {
	return registry.NewRepository(repo)
}

func ProvideRegistryRepositoryPort(r *registry.Repository) registry.RepositoryPort {
	return r
}

func ProvideRegistryUseCases(rep registry.RepositoryPort) *registry.UseCases {
	return registry.NewUseCases(rep)
}

func ProvideRegistryUseCasesPort(uc *registry.UseCases) registry.UseCasesPort {
	return uc
}

func ProvideRegistryHandler(
	server registry.GinEnginePort,
	useCases registry.UseCasesPort,
	cfg registry.ConfigAPIPort,
	middlewares registry.MiddlewaresEnginePort,
) *registry.Handler {
	return registry.NewHandler(useCases, server, cfg, middlewares)
}

func ProvideRegistryConfigAPI(cfg *config.Config) registry.ConfigAPIPort {
	return &cfg.API
}

func ProvideRegistryGormEnginePort(r *pgorm.Repository) registry.GormEnginePort {
	return r
}

func ProvideRegistryGinEnginePort(s *pgin.Server) registry.GinEnginePort {
	return s
}

func ProvideRegistryMiddlewaresEnginePort(m *mwr.Middlewares) registry.MiddlewaresEnginePort {
	return m
}

// RegistrySet expone todos los providers necesarios para Registry.
var RegistrySet = wire.NewSet(
	ProvideRegistryRepository,
	ProvideRegistryRepositoryPort,
	ProvideRegistryUseCases,
	ProvideRegistryUseCasesPort,
	ProvideRegistryHandler,
	ProvideRegistryConfigAPI,
	ProvideRegistryGormEnginePort,
	ProvideRegistryGinEnginePort,
	ProvideRegistryMiddlewaresEnginePort,
)
