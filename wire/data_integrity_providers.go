package wire

import (
	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"

	dashboard "github.com/devpablocristo/ponti-backend/internal/dashboard"
	dataintegrity "github.com/devpablocristo/ponti-backend/internal/data-integrity"
	lot "github.com/devpablocristo/ponti-backend/internal/lot"
	project "github.com/devpablocristo/ponti-backend/internal/project"
	report "github.com/devpablocristo/ponti-backend/internal/report"
	supply "github.com/devpablocristo/ponti-backend/internal/supply"
	workorder "github.com/devpablocristo/ponti-backend/internal/work-order"
)

// ProvideDataIntegrityUseCases construye los casos de uso de dataintegrity.
func ProvideDataIntegrityUseCases(
	dashboardRepo dataintegrity.DashboardRepositoryPort,
	workOrderRepo dataintegrity.WorkOrderRepositoryPort,
	reportRepo dataintegrity.ReportRepositoryPort,
	supplyRepo dataintegrity.SupplyRepositoryPort,
	projectRepo dataintegrity.ProjectRepositoryPort,
	lotRepo dataintegrity.LotRepositoryPort,
) *dataintegrity.UseCases {
	return dataintegrity.NewUseCases(
		dashboardRepo,
		workOrderRepo,
		reportRepo,
		supplyRepo,
		projectRepo,
		lotRepo,
	)
}

// ProvideDataIntegrityUseCasesPort adapta *dataintegrity.UseCases a la interfaz dataintegrity.UseCasesPort.
func ProvideDataIntegrityUseCasesPort(uc *dataintegrity.UseCases) dataintegrity.UseCasesPort {
	return uc
}

// ProvideDataIntegrityHandler construye el handler HTTP para Data Integrity.
func ProvideDataIntegrityHandler(
	server dataintegrity.GinEnginePort,
	useCases dataintegrity.UseCasesPort,
	cfg dataintegrity.ConfigAPIPort,
	middlewares dataintegrity.MiddlewaresEnginePort,
) *dataintegrity.Handler {
	return dataintegrity.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideDataIntegrityConfigAPI extrae la configuración específica de API para Data Integrity.
func ProvideDataIntegrityConfigAPI(cfg *config.Config) dataintegrity.ConfigAPIPort {
	return &cfg.API
}

// ProvideDataIntegrityGinEnginePort adapta *pgin.Server a dataintegrity.GinEnginePort.
func ProvideDataIntegrityGinEnginePort(s *pgin.Server) dataintegrity.GinEnginePort {
	return s
}

// ProvideDataIntegrityMiddlewaresEnginePort adapta *mwr.Middlewares a dataintegrity.MiddlewaresEnginePort.
func ProvideDataIntegrityMiddlewaresEnginePort(m *mwr.Middlewares) dataintegrity.MiddlewaresEnginePort {
	return m
}

// ProvideDataIntegrityDashboardRepositoryPort adapta dashboard.RepositoryPort.
func ProvideDataIntegrityDashboardRepositoryPort(r dashboard.RepositoryPort) dataintegrity.DashboardRepositoryPort {
	return r
}

// ProvideDataIntegrityWorkOrderRepositoryPort adapta workorder.RepositoryPort.
// workorder.RepositoryPort ya incluye GetRawDirectCost, único método requerido por
// la interfaz mínima del módulo data-integrity.
func ProvideDataIntegrityWorkOrderRepositoryPort(r workorder.RepositoryPort) dataintegrity.WorkOrderRepositoryPort {
	return r
}

// ProvideDataIntegrityReportRepositoryPort recibe el repo concreto porque GetRawNetIncome
// vive en *report.ReportRepository pero no se expone en report.ReportRepositoryPort para
// no contaminar la interfaz pública con métodos exclusivos de data-integrity.
func ProvideDataIntegrityReportRepositoryPort(r *report.ReportRepository) dataintegrity.ReportRepositoryPort {
	return r
}

// ProvideDataIntegritySupplyRepositoryPort recibe el repo concreto (ver razón en ReportRepositoryPort).
func ProvideDataIntegritySupplyRepositoryPort(r *supply.Repository) dataintegrity.SupplyRepositoryPort {
	return r
}

// ProvideDataIntegrityProjectRepositoryPort recibe el repo concreto (ver razón en ReportRepositoryPort).
func ProvideDataIntegrityProjectRepositoryPort(r *project.Repository) dataintegrity.ProjectRepositoryPort {
	return r
}

// ProvideDataIntegrityLotRepositoryPort recibe el repo concreto (ver razón en ReportRepositoryPort).
func ProvideDataIntegrityLotRepositoryPort(r *lot.Repository) dataintegrity.LotRepositoryPort {
	return r
}

// DataIntegritySet expone todos los providers necesarios para Data Integrity.
var DataIntegritySet = wire.NewSet(
	ProvideDataIntegrityUseCases,
	ProvideDataIntegrityUseCasesPort,
	ProvideDataIntegrityHandler,
	ProvideDataIntegrityConfigAPI,
	ProvideDataIntegrityGinEnginePort,
	ProvideDataIntegrityMiddlewaresEnginePort,
	ProvideDataIntegrityDashboardRepositoryPort,
	ProvideDataIntegrityWorkOrderRepositoryPort,
	ProvideDataIntegrityReportRepositoryPort,
	ProvideDataIntegritySupplyRepositoryPort,
	ProvideDataIntegrityProjectRepositoryPort,
	ProvideDataIntegrityLotRepositoryPort,
)
