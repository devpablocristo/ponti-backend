package wire

import (
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	mwr "github.com/devpablocristo/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/pkg/http/servers/gin"

	dashboard "github.com/devpablocristo/ponti-backend/internal/dashboard"
	dataintegrity "github.com/devpablocristo/ponti-backend/internal/data-integrity"
	lot "github.com/devpablocristo/ponti-backend/internal/lot"
	report "github.com/devpablocristo/ponti-backend/internal/report"
	stock "github.com/devpablocristo/ponti-backend/internal/stock"
	workorder "github.com/devpablocristo/ponti-backend/internal/work-order"
)

// ProvideDataIntegrityUseCases construye los casos de uso de dataintegrity
func ProvideDataIntegrityUseCases(
	workOrderRepo dataintegrity.WorkOrderRepositoryPort,
	dashboardRepo dataintegrity.DashboardRepositoryPort,
	lotRepo dataintegrity.LotRepositoryPort,
	reportRepo dataintegrity.ReportRepositoryPort,
	stockRepo dataintegrity.StockRepositoryPort,
) *dataintegrity.UseCases {
	return dataintegrity.NewUseCases(
		workOrderRepo,
		dashboardRepo,
		lotRepo,
		reportRepo,
		stockRepo,
	)
}

// ProvideDataIntegrityUseCasesPort adapta *dataintegrity.UseCases a la interfaz dataintegrity.UseCasesPort
func ProvideDataIntegrityUseCasesPort(uc *dataintegrity.UseCases) dataintegrity.UseCasesPort {
	return uc
}

// ProvideDataIntegrityHandler construye el handler HTTP para Data Integrity
func ProvideDataIntegrityHandler(
	server dataintegrity.GinEnginePort,
	useCases dataintegrity.UseCasesPort,
	cfg dataintegrity.ConfigAPIPort,
	middlewares dataintegrity.MiddlewaresEnginePort,
) *dataintegrity.Handler {
	return dataintegrity.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideDataIntegrityConfigAPI extrae la configuración específica de API para Data Integrity
func ProvideDataIntegrityConfigAPI(cfg *config.Config) dataintegrity.ConfigAPIPort {
	return &cfg.API
}

// ProvideDataIntegrityGinEnginePort adapta *pgin.Server a dataintegrity.GinEnginePort
func ProvideDataIntegrityGinEnginePort(s *pgin.Server) dataintegrity.GinEnginePort {
	return s
}

// ProvideDataIntegrityMiddlewaresEnginePort adapta *mwr.Middlewares a dataintegrity.MiddlewaresEnginePort
func ProvideDataIntegrityMiddlewaresEnginePort(m *mwr.Middlewares) dataintegrity.MiddlewaresEnginePort {
	return m
}

// ProvideDataIntegrityWorkOrderRepositoryPort adapta workorder.RepositoryPort a dataintegrity.WorkOrderRepositoryPort.
func ProvideDataIntegrityWorkOrderRepositoryPort(r workorder.RepositoryPort) dataintegrity.WorkOrderRepositoryPort {
	return r
}

// ProvideDataIntegrityDashboardRepositoryPort adapta dashboard.RepositoryPort a dataintegrity.DashboardRepositoryPort
func ProvideDataIntegrityDashboardRepositoryPort(r dashboard.RepositoryPort) dataintegrity.DashboardRepositoryPort {
	return r
}

// ProvideDataIntegrityLotRepositoryPort adapta lot.RepositoryPort a dataintegrity.LotRepositoryPort
func ProvideDataIntegrityLotRepositoryPort(r lot.RepositoryPort) dataintegrity.LotRepositoryPort {
	return r
}

// ProvideDataIntegrityReportRepositoryPort adapta report.ReportRepositoryPort a dataintegrity.ReportRepositoryPort
func ProvideDataIntegrityReportRepositoryPort(r report.ReportRepositoryPort) dataintegrity.ReportRepositoryPort {
	return r
}

// ProvideDataIntegrityStockRepositoryPort adapta stock.RepositoryPort a dataintegrity.StockRepositoryPort
func ProvideDataIntegrityStockRepositoryPort(r stock.RepositoryPort) dataintegrity.StockRepositoryPort {
	return r
}

// DataIntegritySet expone todos los providers necesarios para Data Integrity
var DataIntegritySet = wire.NewSet(
	ProvideDataIntegrityUseCases,
	ProvideDataIntegrityUseCasesPort,
	ProvideDataIntegrityHandler,
	ProvideDataIntegrityConfigAPI,
	ProvideDataIntegrityGinEnginePort,
	ProvideDataIntegrityMiddlewaresEnginePort,
	ProvideDataIntegrityWorkOrderRepositoryPort,
	ProvideDataIntegrityDashboardRepositoryPort,
	ProvideDataIntegrityLotRepositoryPort,
	ProvideDataIntegrityReportRepositoryPort,
	ProvideDataIntegrityStockRepositoryPort,
)
