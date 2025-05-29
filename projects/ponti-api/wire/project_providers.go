package wire

import (
	"github.com/google/wire"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	gin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"

	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

// Gorm
func ProvideProjectGormEnginePort(r *gorm.Repository) project.GormEnginePort {
	return r
}

func ProvideProjectRepository(repo project.GormEnginePort) *project.Repository {
	return project.NewRepository(repo)
}

func ProvideProjectRepositoryPort(r *project.Repository) project.RepositoryPort {
	return r
}



// Suggester
func ProvideProjectSuggesterEnginePort(s *sug.Suggester) project.SuggesterEnginePort {
	return s
}

func ProvideProjectSuggester(suge project.SuggesterEnginePort, table, column string) *project.Suggester {
	return project.NewSuggester(suge, table, column)
}

func ProvideProjectSuggesterPort(s *sug.Suggester) project.SuggesterEnginePort {
	return s
}



// Http Server
func ProvideProjectGinEnginePort(s *gin.Server) project.GinEnginePort {
	return s
}



// ProvideProjectUseCases agrupa repos y usecases de otros dominios en project.UseCases.
func ProvideProjectUseCases(
	rep project.RepositoryPort,
	sugPort project.SuggesterPort,
	cusPort project.CustomerUseCasesPort,
	mgrPort project.ManagerUseCasesPort,
	invPort project.InvestorsUseCasesPort,
	fldPort project.FieldUseCasesPort,
) *project.UseCases {
	return project.NewUseCases(rep, sugPort, cusPort, mgrPort, invPort, fldPort)
}

// ProvideProjectUseCasesPort adapta *project.UseCases a project.UseCasesPort.
func ProvideProjectUseCasesPort(u *project.UseCases) project.UseCasesPort {
	return u
}

// ProvideProjectCustomerUseCasesPort adapta *customer.UseCases a project.CustomerUseCasesPort.
func ProvideProjectCustomerUseCasesPort(c *customer.UseCases) project.CustomerUseCasesPort {
	return c
}

// ProvideProjectManagerUseCasesPort adapta *manager.UseCases a project.ManagerUseCasesPort.
func ProvideProjectManagerUseCasesPort(m *manager.UseCases) project.ManagerUseCasesPort {
	return m
}

// ProvideProjectInvestorsUseCasesPort adapta *investor.UseCases a project.InvestorsUseCasesPort.
func ProvideProjectInvestorsUseCasesPort(i *investor.UseCases) project.InvestorsUseCasesPort {
	return i
}

// ProvideProjectFieldUseCasesPort adapta *field.UseCases a project.FieldUseCasesPort.
func ProvideProjectFieldUseCasesPort(f *field.UseCases) project.FieldUseCasesPort {
	return f
}



// ProvideProjectMiddlewaresEnginePort adapta *mwr.Middlewares a project.MiddlewaresEnginePort.
func ProvideProjectMiddlewaresEnginePort(m *mwr.Middlewares) project.MiddlewaresEnginePort {
	return m
}

func ProvideProjectHandler(
	server project.GinEnginePort,
	useCases project.UseCasesPort,
	cfg project.ConfigAPIPort,
	middlewares project.MiddlewaresEnginePort,
) *project.Handler {
	return project.NewHandler(useCases, server, cfg, middlewares)
}

// ProjectSet expone todos los providers necesarios para Project.
var ProjectSet = wire.NewSet(
	ProvideProjectRepository,
	ProvideProjectRepositoryPort,
	ProvideProjectUseCases,
	ProvideProjectUseCasesPort,
	ProvideProjectHandler,
	ProvideProjectSuggesterPort,
	ProvideProjectCustomerUseCasesPort,
	ProvideProjectManagerUseCasesPort,
	ProvideProjectInvestorsUseCasesPort,
	ProvideProjectFieldUseCasesPort,
	ProvideProjectGormEnginePort,
	ProvideProjectGinEnginePort,
	ProvideProjectMiddlewaresEnginePort,
)
