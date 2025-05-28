// File: wire/project_wire.go
package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	project "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project"
)

// ProvideProjectRepository crea la implementación concreta de project.Repository.
func ProvideProjectRepository(repo project.GormEnginePort) *project.Repository {
	return project.NewRepository(repo)
}

// ProvideProjectUseCases agrupa repos y usecases de otros dominios en project.UseCases.
func ProvideProjectUseCases(
	rep project.RepositoryPort,
	suggester project.SuggesterPort,
	cus customer.UseCases,
	mgr manager.UseCases,
	inv investor.UseCases,
	fld field.UseCases,
	lot lot.UseCases,
) *project.UseCases {
	return project.NewUseCases(rep, suggester, cus, mgr, inv, fld, lot)
}

// ProvideProjectHandler construye el handler HTTP para Project.
func ProvideProjectHandler(
	server project.GinServerPort,
	UseCases project.UseCasesPort,
	config project.ConfigAPIPort,
	middlewares project.MiddlewaresPort,
) *project.Handler {
	return project.NewHandler(UseCases, server, config, middlewares)
}

// ProjectSet expone todos los providers y bindings necesarios para Project.
var ProjectSet = wire.NewSet(
	ProvideProjectRepository,
	ProvideProjectUseCases,
	ProvideProjectHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(project.RepositoryPort), new(*project.Repository)),
	wire.Bind(new(project.UseCasesPort), new(*project.UseCases)),
	wire.Bind(new(project.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(project.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(project.MiddlewaresPort), new(*mwr.Middlewares)),
)
