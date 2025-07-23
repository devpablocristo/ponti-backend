package wire

import (
	"github.com/google/wire"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
	workorder "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder"
)

// ProvideWorkOrderRepository crea la implementación concreta de workorder.Repository.
func ProvideWorkOrderRepository(repo workorder.GormEngine) *workorder.Repository {
	return workorder.NewRepository(repo)
}

// ProvideWorkOrderRepositoryPort adapta *workorder.Repository a la interfaz workorder.RepositoryPort.
func ProvideWorkOrderRepositoryPort(r *workorder.Repository) workorder.RepositoryPort {
	return r
}

// ProvideWorkOrderUseCases agrupa repositorios en workorder.UseCases.
func ProvideWorkOrderUseCases(
	repo workorder.RepositoryPort,
) *workorder.UseCases {
	return workorder.NewUseCases(repo)
}

// ProvideWorkOrderUseCasesPort adapta *workorder.UseCases a la interfaz workorder.UseCasesPort.
func ProvideWorkOrderUseCasesPort(uc *workorder.UseCases) workorder.UseCasesPort {
	return uc
}

// ProvideWorkOrderHandler construye el handler HTTP para WorkOrder.
func ProvideWorkOrderHandler(
	server workorder.GinEnginePort,
	useCases workorder.UseCasesPort,
	cfg workorder.ConfigAPIPort,
	middlewares workorder.MiddlewaresEnginePort,
) *workorder.Handler {
	return workorder.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideWorkOrderConfigAPI extrae la configuración específica de API para WorkOrder.
func ProvideWorkOrderConfigAPI(cfg *config.Config) workorder.ConfigAPIPort {
	return &cfg.API
}

// ProvideWorkOrderGormEnginePort adapta *pgorm.Repository a workorder.GormEngine.
func ProvideWorkOrderGormEnginePort(r *pgorm.Repository) workorder.GormEngine {
	return r
}

// ProvideWorkOrderGinEnginePort adapta *pgin.Server a workorder.GinEnginePort.
func ProvideWorkOrderGinEnginePort(s *pgin.Server) workorder.GinEnginePort {
	return s
}

// ProvideWorkOrderMiddlewaresEnginePort adapta *mwr.Middlewares a workorder.MiddlewaresEnginePort.
func ProvideWorkOrderMiddlewaresEnginePort(m *mwr.Middlewares) workorder.MiddlewaresEnginePort {
	return m
}

// WorkOrderSet expone todos los providers necesarios para WorkOrder.
var WorkOrderSet = wire.NewSet(
	ProvideWorkOrderRepository,
	ProvideWorkOrderRepositoryPort,
	ProvideWorkOrderUseCases,
	ProvideWorkOrderUseCasesPort,
	ProvideWorkOrderHandler,
	ProvideWorkOrderConfigAPI,
	ProvideWorkOrderGormEnginePort,
	ProvideWorkOrderGinEnginePort,
	ProvideWorkOrderMiddlewaresEnginePort,
)
