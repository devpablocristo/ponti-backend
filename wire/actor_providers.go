package wire

import (
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	actor "github.com/devpablocristo/ponti-backend/internal/actor"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

func ProvideActorRepository(repo actor.GormEnginePort) *actor.Repository {
	return actor.NewRepository(repo)
}

func ProvideActorRepositoryPort(r *actor.Repository) actor.RepositoryPort {
	return r
}

func ProvideActorUseCases(repo actor.RepositoryPort) *actor.UseCases {
	return actor.NewUseCases(repo)
}

func ProvideActorUseCasesPort(uc *actor.UseCases) actor.UseCasesPort {
	return uc
}

func ProvideActorHandler(
	server actor.GinEnginePort,
	useCases actor.UseCasesPort,
	cfg actor.ConfigAPIPort,
	middlewares actor.MiddlewaresEnginePort,
) *actor.Handler {
	return actor.NewHandler(useCases, server, cfg, middlewares)
}

func ProvideActorConfigAPI(cfg *config.Config) actor.ConfigAPIPort {
	return &cfg.API
}

func ProvideActorGormEnginePort(r *pgorm.Repository) actor.GormEnginePort {
	return r
}

func ProvideActorGinEnginePort(s *pgin.Server) actor.GinEnginePort {
	return s
}

func ProvideActorMiddlewaresEnginePort(m *mwr.Middlewares) actor.MiddlewaresEnginePort {
	return m
}

var ActorSet = wire.NewSet(
	ProvideActorRepository,
	ProvideActorRepositoryPort,
	ProvideActorUseCases,
	ProvideActorUseCasesPort,
	ProvideActorHandler,
	ProvideActorConfigAPI,
	ProvideActorGormEnginePort,
	ProvideActorGinEnginePort,
	ProvideActorMiddlewaresEnginePort,
)
