package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	unit "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit" // Corrected import for unit
)

// --- GORM & REPO ---
func ProvideUnitGormEnginePort(r *gormpkg.Repository) unit.GormEnginePort {
	return r
}
func ProvideUnitRepository(r unit.GormEnginePort) *unit.Repository {
	return unit.NewRepository(r)
}
func ProvideUnitRepositoryPort(repo *unit.Repository) unit.RepositoryPort {
	return repo
}

// --- CONFIG API ---
func ProvideUnitConfigAPI(c *cfg.Config) unit.ConfigAPIPort {
	return &c.API
}

// --- HTTP & MIDDLEWARE ---
func ProvideUnitGinEnginePort(s *pgin.Server) unit.GinEnginePort {
	return s
}
func ProvideUnitMiddlewaresEnginePort(m *mwr.Middlewares) unit.MiddlewaresEnginePort {
	return m
}

// --- USE CASES ---
func ProvideUnitUseCases(rep unit.RepositoryPort) *unit.UseCases {
	return unit.NewUseCases(rep)
}

// Corrected line: Changed *unit.UseCasesPort to unit.UseCasesPort
func ProvideUnitUseCasesPort(u *unit.UseCases) unit.UseCasesPort {
	return u
}

// --- HANDLER ---
func ProvideUnitHandler(
	server unit.GinEnginePort,
	ucs unit.UseCasesPort,
	cfg unit.ConfigAPIPort,
	mws unit.MiddlewaresEnginePort,
) *unit.Handler {
	return unit.NewHandler(ucs, server, cfg, mws)
}

// --- WIRE SET ---
var UnitSet = wire.NewSet(
	ProvideUnitGormEnginePort,
	ProvideUnitRepository,
	ProvideUnitRepositoryPort,
	ProvideUnitConfigAPI,
	ProvideUnitGinEnginePort,
	ProvideUnitMiddlewaresEnginePort,

	ProvideUnitUseCases,
	ProvideUnitUseCasesPort,
	ProvideUnitHandler,
)
