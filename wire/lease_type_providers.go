package wire

import (
	"github.com/google/wire"

	cfg "github.com/devpablocristo/ponti-backend/cmd/config"
	leasetype "github.com/devpablocristo/ponti-backend/internal/lease-type"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	gormpkg "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

// ProvideLeaseTypeGormEnginePort ...
func ProvideLeaseTypeGormEnginePort(r *gormpkg.Repository) leasetype.GormEnginePort {
	return r
}

// ProvideLeaseTypeRepository ...
func ProvideLeaseTypeRepository(r leasetype.GormEnginePort) *leasetype.Repository {
	return leasetype.NewRepository(r)
}

// ProvideLeaseTypeRepositoryPort ...
func ProvideLeaseTypeRepositoryPort(repo *leasetype.Repository) leasetype.RepositoryPort {
	return repo
}

// ProvideLeaseTypeConfigAPI ...
func ProvideLeaseTypeConfigAPI(c *cfg.Config) leasetype.ConfigAPIPort {
	return &c.API
}

// ProvideLeaseTypeGinEnginePort ...
func ProvideLeaseTypeGinEnginePort(s *pgin.Server) leasetype.GinEnginePort {
	return s
}

// ProvideLeaseTypeMiddlewaresEnginePort ...
func ProvideLeaseTypeMiddlewaresEnginePort(m *mwr.Middlewares) leasetype.MiddlewaresEnginePort {
	return m
}

// ProvideLeaseTypeUseCases ...
func ProvideLeaseTypeUseCases(rep leasetype.RepositoryPort) *leasetype.UseCases {
	return leasetype.NewUseCases(rep)
}

// ProvideLeaseTypeUseCasesPort ...
func ProvideLeaseTypeUseCasesPort(u *leasetype.UseCases) leasetype.UseCasesPort {
	return u
}

// ProvideLeaseTypeHandler ...
func ProvideLeaseTypeHandler(
	server leasetype.GinEnginePort,
	ucs leasetype.UseCasesPort,
	cfg leasetype.ConfigAPIPort,
	mws leasetype.MiddlewaresEnginePort,
) *leasetype.Handler {
	return leasetype.NewHandler(ucs, server, cfg, mws)
}

// LeaseTypeSet ...
var LeaseTypeSet = wire.NewSet(
	ProvideLeaseTypeGormEnginePort,
	ProvideLeaseTypeRepository,
	ProvideLeaseTypeRepositoryPort,
	ProvideLeaseTypeConfigAPI,
	ProvideLeaseTypeGinEnginePort,
	ProvideLeaseTypeMiddlewaresEnginePort,

	ProvideLeaseTypeUseCases,
	ProvideLeaseTypeUseCasesPort,
	ProvideLeaseTypeHandler,
)
