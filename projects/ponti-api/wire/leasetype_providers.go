package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	leasetype "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype"
)

// --- GORM & REPO ---
func ProvideLeaseTypeGormEnginePort(r *gormpkg.Repository) leasetype.GormEnginePort {
	return r
}
func ProvideLeaseTypeRepository(r leasetype.GormEnginePort) *leasetype.Repository {
	return leasetype.NewRepository(r)
}
func ProvideLeaseTypeRepositoryPort(repo *leasetype.Repository) leasetype.RepositoryPort {
	return repo
}

// --- CONFIG API ---
func ProvideLeaseTypeConfigAPI(c *cfg.Config) leasetype.ConfigAPIPort {
	return &c.API
}

// --- HTTP & MIDDLEWARE ---
func ProvideLeaseTypeGinEnginePort(s *pgin.Server) leasetype.GinEnginePort {
	return s
}
func ProvideLeaseTypeMiddlewaresEnginePort(m *mwr.Middlewares) leasetype.MiddlewaresEnginePort {
	return m
}

// --- USE CASES ---
func ProvideLeaseTypeUseCases(rep leasetype.RepositoryPort) *leasetype.UseCases {
	return leasetype.NewUseCases(rep)
}

func ProvideLeaseTypeUseCasesPort(u *leasetype.UseCases) leasetype.UseCasesPort {
	return u
}

// --- HANDLER ---
func ProvideLeaseTypeHandler(
	server leasetype.GinEnginePort,
	ucs leasetype.UseCasesPort,
	cfg leasetype.ConfigAPIPort,
	mws leasetype.MiddlewaresEnginePort,
) *leasetype.Handler {
	return leasetype.NewHandler(ucs, server, cfg, mws)
}

// --- WIRE SET ---
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
