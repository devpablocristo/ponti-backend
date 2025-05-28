package wire

import (
	"github.com/google/wire"

	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	cfg "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"

	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
)

// ProvideFieldRepository crea la implementación concreta de field.Repository.
func ProvideFieldRepository(repo field.GormEnginePort) *field.Repository {
	return field.NewRepository(repo)
}

// ProvideFieldUseCases agrupa repositorios y servicios relacionados en field.UseCases.
func ProvideFieldUseCases(
	rep field.RepositoryPort,
	lotUC lot.UseCasesPort,
) *field.UseCases {
	return field.NewUseCases(rep, lotUC)
}

// ProvideFieldHandler construye el handler HTTP para Field.
func ProvideFieldHandler(
	server field.GinServerPort,
	UseCases field.UseCasesPort,
	config field.ConfigAPIPort,
	middlewares field.MiddlewaresPort,
) *field.Handler {
	return field.NewHandler(UseCases, server, config, middlewares)
}

// FieldSet expone todos los providers y bindings necesarios para Field.
var FieldSet = wire.NewSet(
	ProvideFieldRepository,
	ProvideFieldUseCases,
	ProvideFieldHandler,

	// Bindings de interfaces a implementaciones concretas
	wire.Bind(new(field.RepositoryPort), new(*field.Repository)),
	wire.Bind(new(field.UseCasesPort), new(*field.UseCases)),
	wire.Bind(new(field.GinServerPort), new(*ginsrv.Server)),
	wire.Bind(new(field.ConfigAPIPort), new(*cfg.API)),
	wire.Bind(new(field.MiddlewaresPort), new(*mwr.Middlewares)),
)
