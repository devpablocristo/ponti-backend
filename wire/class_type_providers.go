package wire

import (
	"github.com/google/wire"

	cfg "github.com/devpablocristo/ponti-backend/cmd/config"
	classtype "github.com/devpablocristo/ponti-backend/internal/class-type" // Corrected import for classtype
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	gormpkg "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

// ProvideClassTypeGormEnginePort ...
func ProvideClassTypeGormEnginePort(r *gormpkg.Repository) classtype.GormEnginePort {
	return r
}
func ProvideClassTypeRepository(r classtype.GormEnginePort) *classtype.Repository {
	return classtype.NewRepository(r)
}
func ProvideClassTypeRepositoryPort(repo *classtype.Repository) classtype.RepositoryPort {
	return repo
}

// ProvideClassTypeConfigAPI ...
func ProvideClassTypeConfigAPI(c *cfg.Config) classtype.ConfigAPIPort {
	return &c.API
}

// ProvideClassTypeGinEnginePort ...
func ProvideClassTypeGinEnginePort(s *pgin.Server) classtype.GinEnginePort {
	return s
}
func ProvideClassTypeMiddlewaresEnginePort(m *mwr.Middlewares) classtype.MiddlewaresEnginePort {
	return m
}

// ProvideClassTypeUseCases ...
func ProvideClassTypeUseCases(rep classtype.RepositoryPort) *classtype.UseCases {
	return classtype.NewUseCases(rep)
}

// ProvideClassTypeUseCasesPort ...
func ProvideClassTypeUseCasesPort(u *classtype.UseCases) classtype.UseCasesPort {
	return u
}

// ProvideClassTypeHandler ...
func ProvideClassTypeHandler(
	server classtype.GinEnginePort,
	ucs classtype.UseCasesPort,
	cfg classtype.ConfigAPIPort,
	mws classtype.MiddlewaresEnginePort,
) *classtype.Handler {
	return classtype.NewHandler(ucs, server, cfg, mws)
}

// ClassTypeSet ...
var ClassTypeSet = wire.NewSet(
	ProvideClassTypeGormEnginePort,
	ProvideClassTypeRepository,
	ProvideClassTypeRepositoryPort,
	ProvideClassTypeConfigAPI,
	ProvideClassTypeGinEnginePort,
	ProvideClassTypeMiddlewaresEnginePort,

	ProvideClassTypeUseCases,
	ProvideClassTypeUseCasesPort,
	ProvideClassTypeHandler,
)
