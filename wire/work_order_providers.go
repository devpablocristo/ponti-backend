package wire

import (
	"os"
	"path/filepath"

	"github.com/google/wire"

	pgorm "github.com/devpablocristo/ponti-backend/pkg/databases/sql/gorm"
	pkgexcel "github.com/devpablocristo/ponti-backend/pkg/files-io/excel/excelize"
	mwr "github.com/devpablocristo/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/pkg/http/servers/gin"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	workorder "github.com/devpablocristo/ponti-backend/internal/work-order"
	workOrderExcel "github.com/devpablocristo/ponti-backend/internal/work-order/excel"
)

// ProvideWorkOrderRepository crea la implementación concreta de workorder.Repository.
func ProvideWorkOrderRepository(repo workorder.GormEngine) *workorder.Repository {
	return workorder.NewRepository(repo)
}

// ProvideWorkOrderRepositoryPort adapta *workorder.Repository a la interfaz workorder.RepositoryPort.
func ProvideWorkOrderRepositoryPort(r *workorder.Repository) workorder.RepositoryPort {
	return r
}

// Crea el engine de Excel ya configurado
func ProvidePkgExcelService() (*pkgexcel.Service, error) {
	fp := filepath.Join(os.TempDir(), workOrderExcel.DefaultFilename)
	write := true
	return pkgexcel.Bootstrap(
		fp,
		workOrderExcel.SheetName,
		workOrderExcel.DateFormat,
		&write,
		workOrderExcel.ColumnWidths,
	)
}

// bindea el engine como la interfaz XLSXEnginePort
func ProvideXLSXEnginePort(s *pkgexcel.Service) workorder.XLSXEnginePort {
	return s
}

// Crea el adaptador de exportación que usa el engine
func ProvideExporterPort(eng workorder.XLSXEnginePort) workorder.ExporterAdapterPort {
	return workorder.NewExcelExporter(eng)
}

// ProvideWorkOrderUseCases agrupa repositorios en workorder.UseCases.
func ProvideWorkOrderUseCases(repo workorder.RepositoryPort, exp workorder.ExporterAdapterPort) *workorder.UseCases {
	return workorder.NewUseCases(repo, exp)
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
	ProvidePkgExcelService,
	ProvideExporterPort,
	ProvideXLSXEnginePort,
)
