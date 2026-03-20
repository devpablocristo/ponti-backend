package wire

import (
	gormpkg "github.com/devpablocristo/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/devpablocristo/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/pkg/http/servers/gin"
	"github.com/google/wire"

	cfg "github.com/devpablocristo/ponti-backend/cmd/config"
	commercialization "github.com/devpablocristo/ponti-backend/internal/commercialization"
)

// ---- GORM -----

func ProvideCommercializationGormEnginePort(r *gormpkg.Repository) commercialization.GormEnginePort {
	return r
}

func ProvideCommercializationRepository(r commercialization.GormEnginePort) *commercialization.Repository {
	return commercialization.NewRepository(r)
}

func ProvideCommercializationRepositoryPort(r *commercialization.Repository) commercialization.RepositoryPort {
	return r
}

// ---- CONFIG API ----

func ProvideCommercializationConfigAPI(c *cfg.Config) commercialization.ConfigAPIPort {
	return &c.API
}

// ---- HTTP & MIDDLEWARE ---

func ProvideCommercializationGinEnginePort(s *pgin.Server) commercialization.GinEnginePort {
	return s
}

func ProvideCommercializationMiddlewaresEnginePort(m *mwr.Middlewares) commercialization.MiddlewaresEnginePort {
	return m
}

// ---- USE CASES ----

func ProvideCommercializationUseCases(r commercialization.RepositoryPort) *commercialization.UseCases {
	return commercialization.NewUseCases(r)
}

func ProvideCommercializationUseCasePort(u *commercialization.UseCases) commercialization.UseCasePort {
	return u
}

// ---- HANDLER ----

func ProvideCommercializationHandler(
	server commercialization.GinEnginePort,
	ucs commercialization.UseCasePort,
	cfg commercialization.ConfigAPIPort,
	mws commercialization.MiddlewaresEnginePort,
) *commercialization.Handler {
	return commercialization.NewHandler(ucs, server, cfg, mws)
}

var CommercializationSet = wire.NewSet(
	ProvideCommercializationGormEnginePort,
	ProvideCommercializationRepository,
	ProvideCommercializationRepositoryPort,

	ProvideCommercializationConfigAPI,

	ProvideCommercializationGinEnginePort,
	ProvideCommercializationMiddlewaresEnginePort,

	ProvideCommercializationUseCases,
	ProvideCommercializationUseCasePort,

	ProvideCommercializationHandler,
)
