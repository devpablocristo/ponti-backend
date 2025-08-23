package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	unit "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit" // Corrected import for unit
)

// ProvideUnitGormEnginePort ...
func ProvideUnitGormEnginePort(r *gormpkg.Repository) unit.GormEnginePort {
	return r
}

// ProvideUnitRepository ...
func ProvideUnitRepository(r unit.GormEnginePort) *unit.Repository {
	return unit.NewRepository(r)
}

// ProvideUnitRepositoryPort ...
func ProvideUnitRepositoryPort(repo *unit.Repository) unit.RepositoryPort {
	return repo
}

// ProvideUnitConfigAPI ...
func ProvideUnitConfigAPI(c *cfg.Config) unit.ConfigAPIPort {
	return &c.API
}

// ProvideUnitGinEnginePort ...
func ProvideUnitGinEnginePort(s *pgin.Server) unit.GinEnginePort {
	return s
}

// ProvideUnitMiddlewaresEnginePort ...
func ProvideUnitMiddlewaresEnginePort(m *mwr.Middlewares) unit.MiddlewaresEnginePort {
	return m
}

// ProvideUnitUseCases ...
func ProvideUnitUseCases(rep unit.RepositoryPort) *unit.UseCases {
	return unit.NewUseCases(rep)
}

// Corrected line: Changed *unit.UseCasesPort to unit.UseCasesPort
// ProvideUnitUseCasesPort ...
func ProvideUnitUseCasesPort(u *unit.UseCases) unit.UseCasesPort {
	return u
}

// ProvideUnitHandler ...
func ProvideUnitHandler(
	server unit.GinEnginePort,
	ucs unit.UseCasesPort,
	cfg unit.ConfigAPIPort,
	mws unit.MiddlewaresEnginePort,
) *unit.Handler {
	return unit.NewHandler(ucs, server, cfg, mws)
}

// UnitSet ...
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
