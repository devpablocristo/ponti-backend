package wire

import (
	"github.com/google/wire"

	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	actors "github.com/devpablocristo/ponti-backend/internal/actors"
)

// ProvideActorsRepository crea la implementación concreta de actors.Repository.
func ProvideActorsRepository(repo actors.GormEnginePort) *actors.Repository {
	return actors.NewRepository(repo)
}

// ProvideActorsRepositoryPort adapta *actors.Repository a actors.RepositoryPort.
func ProvideActorsRepositoryPort(r *actors.Repository) actors.RepositoryPort {
	return r
}

// ProvideActorsUseCases agrupa repositorios en actors.UseCases.
func ProvideActorsUseCases(repo actors.RepositoryPort) *actors.UseCases {
	return actors.NewUseCases(repo)
}

// ProvideActorsUseCasesPort adapta *actors.UseCases a actors.UseCasesPort.
func ProvideActorsUseCasesPort(uc *actors.UseCases) actors.UseCasesPort {
	return uc
}

// ProvideActorsHandler construye el handler HTTP para Actors.
func ProvideActorsHandler(
	server actors.GinEnginePort,
	useCases actors.UseCasesPort,
	cfg actors.ConfigAPIPort,
	middlewares actors.MiddlewaresEnginePort,
) *actors.Handler {
	return actors.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideActorsConfigAPI extrae la configuración de API para Actors.
func ProvideActorsConfigAPI(cfg *config.Config) actors.ConfigAPIPort {
	return &cfg.API
}

// ProvideActorsGormEnginePort adapta *pgorm.Repository a actors.GormEnginePort.
func ProvideActorsGormEnginePort(r *pgorm.Repository) actors.GormEnginePort {
	return r
}

// ProvideActorsGinEnginePort adapta *pgin.Server a actors.GinEnginePort.
func ProvideActorsGinEnginePort(s *pgin.Server) actors.GinEnginePort {
	return s
}

// ProvideActorsMiddlewaresEnginePort adapta *mwr.Middlewares a actors.MiddlewaresEnginePort.
func ProvideActorsMiddlewaresEnginePort(m *mwr.Middlewares) actors.MiddlewaresEnginePort {
	return m
}

// ActorsSet expone todos los providers necesarios para Actors.
var ActorsSet = wire.NewSet(
	ProvideActorsRepository,
	ProvideActorsRepositoryPort,
	ProvideActorsUseCases,
	ProvideActorsUseCasesPort,
	ProvideActorsHandler,
	ProvideActorsConfigAPI,
	ProvideActorsGormEnginePort,
	ProvideActorsGinEnginePort,
	ProvideActorsMiddlewaresEnginePort,
)
