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

// --- GORM & REPOSITORIO ---
func ProvideProjectGormEnginePort(r *gormpkg.Repository) project.GormEnginePort {
	return r
}
func ProvideProjectRepository(repo project.GormEnginePort) *project.Repository {
	return project.NewRepository(repo)
}
func ProvideProjectRepositoryPort(r *project.Repository) project.RepositoryPort {
	return r
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

// --- USE CASES ---
func ProvideProjectUseCases(
	rep project.RepositoryPort,
	sug project.SuggesterPort,
	cus customer.UseCasesPort,
	mgr manager.UseCasesPort,
	inv investor.UseCasesPort,
	fld field.UseCasesPort,
) *project.UseCases {
	return project.NewUseCases(rep, sug, cus, mgr, inv, fld)
}
func ProvideProjectUseCasesPort(u *project.UseCases) project.UseCasesPort {
	return u
}

// --- HANDLER ---
func ProvideProjectHandler(
	server project.GinEnginePort,
	useCases project.UseCasesPort,
	cfg project.ConfigAPIPort,
	middlewares project.MiddlewaresEnginePort,
) *project.Handler {
	return project.NewHandler(useCases, server, cfg, middlewares)
}

// --------------------------------------------------------------------------------
// Ahora la parte “ad hoc” del Suggester para Project
// --------------------------------------------------------------------------------

// 1) Alias de tipo para desambiguar string
type ProjectTableName string

func ProvideProjectTableName() ProjectTableName { return ProjectTableName("projects") }

type ProjectColumnName string

func ProvideProjectColumnName() ProjectColumnName { return ProjectColumnName("name") }

// 2) Motor común
func ProvideProjectSuggesterEnginePort(s *sug.Suggester) project.SuggesterEnginePort {
	return s
}

// 3) Proveedor específico inyectando table/column
func ProvideProjectSuggester(
	eng project.SuggesterEnginePort,
	table ProjectTableName,
	column ProjectColumnName,
) *project.Suggester {
	return project.NewSuggester(eng, string(table), string(column))
}
func ProvideProjectSuggesterPort(s *project.Suggester) project.SuggesterPort {
	return s
}

// --------------------------------------------------------------------------------
// El set completo
// --------------------------------------------------------------------------------
var ProjectSet = wire.NewSet(
	// GORM & Repo
	ProvideProjectGormEnginePort,
	ProvideProjectRepository,
	ProvideProjectRepositoryPort,

	// Config API
	ProvideProjectConfigAPI,

	// HTTP & Middleware
	ProvideProjectGinEnginePort,
	ProvideProjectMiddlewaresEnginePort,

	// Use Cases
	ProvideProjectUseCases,
	ProvideProjectUseCasesPort,

	// Handler
	ProvideProjectHandler,

	// Suggester “ad hoc”
	ProvideProjectTableName,
	ProvideProjectColumnName,
	ProvideProjectSuggesterEnginePort,
	ProvideProjectSuggester,
	ProvideProjectSuggesterPort,
)
