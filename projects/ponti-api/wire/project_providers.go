package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

// ProvideProjectRepository crea la implementación concreta de project.Repository.
func ProvideProjectRepository(repo project.GormEnginePort) *project.Repository {
	return project.NewRepository(repo)
}

// ProvideProjectRepositoryPort adapta *project.Repository a project.RepositoryPort.
func ProvideProjectRepositoryPort(r *project.Repository) project.RepositoryPort {
	return r
}

// ProvideProjectUseCases agrupa repos y usecases de otros dominios en project.UseCases.
func ProvideProjectUseCases(
	rep project.RepositoryPort,
	sug project.SuggesterPort,
	cus project.CustomerUseCasesPort,
	mgr project.ManagerUseCasesPort,
	inv project.InvestorsUseCasesPort,
	fld project.FieldUseCasesPort,
) *project.UseCases {
	return project.NewUseCases(rep, sug, cus, mgr, inv, fld)
}

// ProvideProjectUseCasesPort adapta *project.UseCases a project.UseCasesPort.
func ProvideProjectUseCasesPort(uc *project.UseCases) project.UseCasesPort {
	return uc
}

// ProvideProjectHandler construye el handler HTTP para Project.
func ProvideProjectHandler(
	server project.GinServerPort,
	useCases project.UseCasesPort,
	cfg project.ConfigAPIPort,
	middlewares project.MiddlewaresPort,
) *project.Handler {
	return project.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideProjectAPIConfig extrae la configuración específica de API para Project.
func ProvideProjectAPIConfig(cfg *config.ConfigSet) project.ConfigAPIPort {
	return &cfg.API
}

// ProvideProjectGormEnginePort adapta *pgorm.Repository a project.GormEnginePort.
func ProvideProjectGormEnginePort(repo *pgorm.Repository) project.GormEnginePort {
	return repo
}

// ProvideProjectGinServerPort adapta *pgin.Server a project.GinServerPort.
func ProvideProjectGinServerPort(srv *pgin.Server) project.GinServerPort {
	return srv
}

// ProvideProjectMiddlewaresPort adapta *mwr.Middlewares a project.MiddlewaresPort.
func ProvideProjectMiddlewaresPort(m *mwr.Middlewares) project.MiddlewaresPort {
	return m
}

// ProvideProjectSuggesterPort permite el paso de un implementador de Suggester.
func ProvideProjectSuggesterPort(s project.SuggesterPort) project.SuggesterPort {
	return s
}

// ProjectSet expone todos los providers necesarios para Project.
var ProjectSet = wire.NewSet(
	ProvideProjectRepository,
	ProvideProjectRepositoryPort,
	ProvideProjectUseCases,
	ProvideProjectUseCasesPort,
	ProvideProjectHandler,
	ProvideProjectAPIConfig,
	ProvideProjectGormEnginePort,
	ProvideProjectGinServerPort,
	ProvideProjectMiddlewaresPort,
	ProvideProjectSuggesterPort,
)
