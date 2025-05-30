package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"

	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

// --- GORM & REPO ---
func ProvideProjectGormEnginePort(r *gormpkg.Repository) project.GormEnginePort {
	return r
}
func ProvideProjectRepository(r project.GormEnginePort) *project.Repository {
	return project.NewRepository(r)
}
func ProvideProjectRepositoryPort(repo *project.Repository) project.RepositoryPort {
	return repo
}

// --- CONFIG API ---
func ProvideProjectConfigAPI(c *cfg.AllConfigs) project.ConfigAPIPort {
	return &c.API
}

// --- HTTP & MIDDLEWARE ---
func ProvideProjectGinEnginePort(s *pgin.Server) project.GinEnginePort {
	return s
}
func ProvideProjectMiddlewaresEnginePort(m *mwr.Middlewares) project.MiddlewaresEnginePort {
	return m
}

// --- SUGGESTER PORT ---
func ProvideProjectSuggesterPort(s *sug.Suggester) project.SuggesterPort {
	return project.NewSuggester(s)
}

// --- USE CASES ---
func ProvideProjectUseCases(
	rep project.RepositoryPort,
	sugg project.SuggesterPort,
	cus customer.UseCasesPort,
	mgr manager.UseCasesPort,
	inv investor.UseCasesPort,
	fld field.UseCasesPort,
) *project.UseCases {
	return project.NewUseCases(rep, sugg, cus, mgr, inv, fld)
}
func ProvideProjectUseCasesPort(u *project.UseCases) project.UseCasesPort {
	return u
}

// --- HANDLER ---
func ProvideProjectHandler(
	server project.GinEnginePort,
	ucs project.UseCasesPort,
	cfg project.ConfigAPIPort,
	mws project.MiddlewaresEnginePort,
) *project.Handler {
	return project.NewHandler(ucs, server, cfg, mws)
}

// --- WIRE SET ---
var ProjectSet = wire.NewSet(
	ProvideProjectGormEnginePort,
	ProvideProjectRepository,
	ProvideProjectRepositoryPort,
	ProvideProjectConfigAPI,
	ProvideProjectGinEnginePort,
	ProvideProjectMiddlewaresEnginePort,

	ProvideProjectSuggesterPort,

	ProvideProjectUseCases,
	ProvideProjectUseCasesPort,
	ProvideProjectHandler,
)
