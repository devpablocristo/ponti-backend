package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	workorder "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder"
)

// ProvideWorkorderRepository crea la implementación concreta de workorder.Repository.
func ProvideWorkorderRepository(repo workorder.GormEngine) *workorder.Repository {
	return workorder.NewRepository(repo)
}

// ProvideWorkorderRepositoryPort adapta *workorder.Repository a la interfaz workorder.RepositoryPort.
func ProvideWorkorderRepositoryPort(r *workorder.Repository) workorder.RepositoryPort {
	return r
}

// ProvideWorkorderUseCases agrupa repositorios en workorder.UseCases.
func ProvideWorkorderUseCases(
	repo workorder.RepositoryPort,
) *workorder.UseCases {
	return workorder.NewUseCases(repo)
}

// ProvideWorkorderUseCasesPort adapta *workorder.UseCases a la interfaz workorder.UseCasesPort.
func ProvideWorkorderUseCasesPort(uc *workorder.UseCases) workorder.UseCasesPort {
	return uc
}

// ProvideWorkorderHandler construye el handler HTTP para Workorder.
func ProvideWorkorderHandler(
	server workorder.GinEnginePort,
	useCases workorder.UseCasesPort,
	cfg workorder.ConfigAPIPort,
	middlewares workorder.MiddlewaresEnginePort,
) *workorder.Handler {
	return workorder.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideWorkorderConfigAPI extrae la configuración específica de API para Workorder.
func ProvideWorkorderConfigAPI(cfg *config.Config) workorder.ConfigAPIPort {
	return &cfg.API
}

// ProvideWorkorderGormEnginePort adapta *pgorm.Repository a workorder.GormEngine.
func ProvideWorkorderGormEnginePort(r *pgorm.Repository) workorder.GormEngine {
	return r
}

// ProvideWorkorderGinEnginePort adapta *pgin.Server a workorder.GinEnginePort.
func ProvideWorkorderGinEnginePort(s *pgin.Server) workorder.GinEnginePort {
	return s
}

// ProvideWorkorderMiddlewaresEnginePort adapta *mwr.Middlewares a workorder.MiddlewaresEnginePort.
func ProvideWorkorderMiddlewaresEnginePort(m *mwr.Middlewares) workorder.MiddlewaresEnginePort {
	return m
}

// WorkorderSet expone todos los providers necesarios para Workorder.
var WorkorderSet = wire.NewSet(
	ProvideWorkorderRepository,
	ProvideWorkorderRepositoryPort,
	ProvideWorkorderUseCases,
	ProvideWorkorderUseCasesPort,
	ProvideWorkorderHandler,
	ProvideWorkorderConfigAPI,
	ProvideWorkorderGormEnginePort,
	ProvideWorkorderGinEnginePort,
	ProvideWorkorderMiddlewaresEnginePort,
)
