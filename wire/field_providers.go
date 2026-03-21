package wire

import (
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"

	field "github.com/devpablocristo/ponti-backend/internal/field"
)

// ProvideFieldRepository crea la implementación concreta de field.Repository.
func ProvideFieldRepository(repo field.GormEnginePort) *field.Repository {
	return field.NewRepository(repo)
}

// ProvideFieldRepositoryPort adapta *field.Repository a la interfaz field.RepositoryPort.
func ProvideFieldRepositoryPort(r *field.Repository) field.RepositoryPort {
	return r
}

// ProvideFieldUseCases agrupa repositorios y servicios relacionados en field.UseCases.
func ProvideFieldUseCases(rep field.RepositoryPort) *field.UseCases {
	return field.NewUseCases(rep)
}

// ProvideFieldUseCasesPort adapta *field.UseCases a la interfaz field.UseCasesPort.
func ProvideFieldUseCasesPort(uc *field.UseCases) field.UseCasesPort {
	return uc
}

// ProvideFieldHandler construye el handler HTTP para Field.
func ProvideFieldHandler(
	server field.GinEnginePort,
	useCases field.UseCasesPort,
	cfg field.ConfigAPIPort,
	middlewares field.MiddlewaresEnginePort,
) *field.Handler {
	return field.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideFieldConfigAPI extrae la configuración específica de API para Field.
func ProvideFieldConfigAPI(cfg *config.Config) field.ConfigAPIPort {
	return &cfg.API
}

// ProvideFieldGormEnginePort adapta *pgorm.Repository a field.GormEnginePort.
func ProvideFieldGormEnginePort(r *pgorm.Repository) field.GormEnginePort {
	return r
}

// ProvideFieldGinEnginePort adapta *pgin.Server a field.GinEnginePort.
func ProvideFieldGinEnginePort(s *pgin.Server) field.GinEnginePort {
	return s
}

// ProvideFieldMiddlewaresEnginePort adapta *mwr.Middlewares a field.MiddlewaresEnginePort.
func ProvideFieldMiddlewaresEnginePort(m *mwr.Middlewares) field.MiddlewaresEnginePort {
	return m
}

// FieldSet expone todos los providers necesarios para Field.
var FieldSet = wire.NewSet(
	ProvideFieldRepository,
	ProvideFieldRepositoryPort,
	ProvideFieldUseCases,
	ProvideFieldUseCasesPort,
	ProvideFieldHandler,
	ProvideFieldConfigAPI,
	ProvideFieldGormEnginePort,
	ProvideFieldGinEnginePort,
	ProvideFieldMiddlewaresEnginePort,
)
