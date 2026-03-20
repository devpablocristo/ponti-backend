package wire

import (
	gormpkg "github.com/devpablocristo/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/devpablocristo/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/pkg/http/servers/gin"
	"github.com/google/wire"

	cfg "github.com/devpablocristo/ponti-backend/cmd/config"
	dollar "github.com/devpablocristo/ponti-backend/internal/dollar"
)

// ---- GORM -----

func ProvideDollarGormEnginePort(r *gormpkg.Repository) dollar.GormEnginePort {
	return r
}

func ProvideDollarRepository(r dollar.GormEnginePort) *dollar.Repository {
	return dollar.NewRepository(r)
}

func ProvideDollarRepositoryPort(r *dollar.Repository) dollar.RepositoryPort {
	return r
}

// ---- CONFIG API ----

func ProvideDollarConfigAPI(c *cfg.Config) dollar.ConfigAPIPort {
	return &c.API
}

// ---- HTTP & MIDDLEWARE ---

func ProvideDollarGinEnginePort(s *pgin.Server) dollar.GinEnginePort {
	return s
}

func ProvideDollarMiddlewaresEnginePort(m *mwr.Middlewares) dollar.MiddlewaresEnginePort {
	return m
}

// ---- USE CASES ----

func ProvideDollarUseCases(r dollar.RepositoryPort) *dollar.UseCases {
	return dollar.NewUseCases(r)
}

func ProvideDollarUseCasePort(u *dollar.UseCases) dollar.UseCasePort {
	return u
}

// ---- HANDLER ----

func ProvideDollarHandler(
	server dollar.GinEnginePort,
	ucs dollar.UseCasePort,
	cfg dollar.ConfigAPIPort,
	mws dollar.MiddlewaresEnginePort,
) *dollar.Handler {
	return dollar.NewHandler(ucs, server, cfg, mws)
}

var DollarSet = wire.NewSet(
	ProvideDollarGormEnginePort,
	ProvideDollarRepository,
	ProvideDollarRepositoryPort,

	ProvideDollarConfigAPI,

	ProvideDollarGinEnginePort,
	ProvideDollarMiddlewaresEnginePort,

	ProvideDollarUseCases,
	ProvideDollarUseCasePort,

	ProvideDollarHandler,
)
