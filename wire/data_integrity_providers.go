package wire

import (
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"

	dashboard "github.com/alphacodinggroup/ponti-backend/internal/dashboard"
	data_integrity "github.com/alphacodinggroup/ponti-backend/internal/data-integrity"
	lot "github.com/alphacodinggroup/ponti-backend/internal/lot"
	report "github.com/alphacodinggroup/ponti-backend/internal/report"
	stock "github.com/alphacodinggroup/ponti-backend/internal/stock"
	workorder "github.com/alphacodinggroup/ponti-backend/internal/work-order"
)

// ProvideDataIntegrityUseCases construye los casos de uso de data_integrity
func ProvideDataIntegrityUseCases(
	workOrderRepo data_integrity.WorkOrderRepositoryPort,
	dashboardRepo data_integrity.DashboardRepositoryPort,
	lotRepo data_integrity.LotRepositoryPort,
	reportRepo data_integrity.ReportRepositoryPort,
	stockRepo data_integrity.StockRepositoryPort,
) *data_integrity.UseCases {
	return data_integrity.NewUseCases(
		workOrderRepo,
		dashboardRepo,
		lotRepo,
		reportRepo,
		stockRepo,
	)
}

// ProvideDataIntegrityUseCasesPort adapta *data_integrity.UseCases a la interfaz data_integrity.UseCasesPort
func ProvideDataIntegrityUseCasesPort(uc *data_integrity.UseCases) data_integrity.UseCasesPort {
	return uc
}

// ProvideDataIntegrityHandler construye el handler HTTP para Data Integrity
func ProvideDataIntegrityHandler(
	server data_integrity.GinEnginePort,
	useCases data_integrity.UseCasesPort,
	cfg data_integrity.ConfigAPIPort,
	middlewares data_integrity.MiddlewaresEnginePort,
) *data_integrity.Handler {
	return data_integrity.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideDataIntegrityConfigAPI extrae la configuración específica de API para Data Integrity
func ProvideDataIntegrityConfigAPI(cfg *config.Config) data_integrity.ConfigAPIPort {
	return &cfg.API
}

// ProvideDataIntegrityGinEnginePort adapta *pgin.Server a data_integrity.GinEnginePort
func ProvideDataIntegrityGinEnginePort(s *pgin.Server) data_integrity.GinEnginePort {
	return s
}

// ProvideDataIntegrityMiddlewaresEnginePort adapta *mwr.Middlewares a data_integrity.MiddlewaresEnginePort
func ProvideDataIntegrityMiddlewaresEnginePort(m *mwr.Middlewares) data_integrity.MiddlewaresEnginePort {
	return m
}

// ProvideDataIntegrityWorkOrderRepositoryPort adapta workorder.RepositoryPort a data_integrity.WorkOrderRepositoryPort.
func ProvideDataIntegrityWorkOrderRepositoryPort(r workorder.RepositoryPort) data_integrity.WorkOrderRepositoryPort {
	return r
}

// ProvideDataIntegrityDashboardRepositoryPort adapta dashboard.RepositoryPort a data_integrity.DashboardRepositoryPort
func ProvideDataIntegrityDashboardRepositoryPort(r dashboard.RepositoryPort) data_integrity.DashboardRepositoryPort {
	return r
}

// ProvideDataIntegrityLotRepositoryPort adapta lot.RepositoryPort a data_integrity.LotRepositoryPort
func ProvideDataIntegrityLotRepositoryPort(r lot.RepositoryPort) data_integrity.LotRepositoryPort {
	return r
}

// ProvideDataIntegrityReportRepositoryPort adapta report.ReportRepositoryPort a data_integrity.ReportRepositoryPort
func ProvideDataIntegrityReportRepositoryPort(r report.ReportRepositoryPort) data_integrity.ReportRepositoryPort {
	return r
}

// ProvideDataIntegrityStockRepositoryPort adapta stock.RepositoryPort a data_integrity.StockRepositoryPort
func ProvideDataIntegrityStockRepositoryPort(r stock.RepositoryPort) data_integrity.StockRepositoryPort {
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
