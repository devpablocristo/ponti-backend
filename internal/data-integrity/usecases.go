// Package dataintegrity implementa controles de coherencia entre módulos.
//
// Filosofía: cada control compara un valor calculado por dos caminos genuinamente
// independientes — la vista SSOT que el usuario ve en pantalla (alimentada por
// v4_report.* / v4_ssot.* / v4_calc.*) vs un cálculo RAW directo contra las tablas
// base (public.*). Si difieren más allá de la tolerancia → ERROR. Si coinciden → OK.
// No hay estados intermedios ni warnings hardcodeados.
package dataintegrity

import (
	"context"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	dashboardDomain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
	supplyDomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

const (
	checkStatusOK    = "OK"
	checkStatusError = "ERROR"

	checkTypeStrong = "STRONG"

	checkSeverityInfo  = "INFO"
	checkSeverityError = "ERROR"
)

// DashboardRepositoryPort expone los valores SSOT que el dashboard muestra al usuario
// (alimentados por v4_report.dashboard_management_balance).
type DashboardRepositoryPort interface {
	GetDashboard(ctx context.Context, filter dashboardDomain.DashboardFilter) (*dashboardDomain.DashboardData, error)
}

// WorkOrderRepositoryPort expone el cálculo RAW de costos directos (∑wo + ∑wi desde tablas base).
type WorkOrderRepositoryPort interface {
	GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error)
}

// ReportRepositoryPort expone el cálculo RAW del ingreso neto (∑lots.tons × cc.net_price).
type ReportRepositoryPort interface {
	GetRawNetIncome(ctx context.Context, projectID int64) (decimal.Decimal, error)
}

// SupplyRepositoryPort expone el cálculo RAW de inversión en insumos
// (∑supply_movements.quantity × supplies.price, filtrado por categoría y tipo de movimiento).
type SupplyRepositoryPort interface {
	GetRawSupplyInvestment(ctx context.Context, projectID int64) (decimal.Decimal, error)
	ListTentativePrices(ctx context.Context, filter supplyDomain.SupplyFilter, limit int) ([]supplyDomain.TentativePriceItem, int64, error)
}

// ProjectRepositoryPort expone el cálculo RAW del costo administrativo total
// (projects.admin_cost × Σlots.hectares).
type ProjectRepositoryPort interface {
	GetRawAdminCostTotal(ctx context.Context, projectID int64) (decimal.Decimal, error)
}

// LotRepositoryPort expone el cálculo RAW del arriendo ejecutado (solo lease_type fijo 3 y 4).
type LotRepositoryPort interface {
	GetRawLeaseExecuted(ctx context.Context, projectID int64) (decimal.Decimal, error)
}

// UseCases orquesta los controles de integridad.
type UseCases struct {
	dashboardRepo DashboardRepositoryPort
	workOrderRepo WorkOrderRepositoryPort
	reportRepo    ReportRepositoryPort
	supplyRepo    SupplyRepositoryPort
	projectRepo   ProjectRepositoryPort
	lotRepo       LotRepositoryPort
}

// NewUseCases crea una nueva instancia de UseCases.
func NewUseCases(
	dashboardRepo DashboardRepositoryPort,
	workOrderRepo WorkOrderRepositoryPort,
	reportRepo ReportRepositoryPort,
	supplyRepo SupplyRepositoryPort,
	projectRepo ProjectRepositoryPort,
	lotRepo LotRepositoryPort,
) *UseCases {
	return &UseCases{
		dashboardRepo: dashboardRepo,
		workOrderRepo: workOrderRepo,
		reportRepo:    reportRepo,
		supplyRepo:    supplyRepo,
		projectRepo:   projectRepo,
		lotRepo:       lotRepo,
	}
}

// defaultTolerance es 1 USD para todos los controles. Los valores SSOT y RAW provienen
// de queries con redondeos numéricos distintos, así que sub-USD no se considera error real.
var defaultTolerance = decimal.NewFromInt(1)

// CheckCostsCoherence orquesta los 5 controles en paralelo.
//
// Single fetch SSOT: trae el DashboardData una vez al inicio (todos los SystemValues
// salen de ahí). Cada control hace después una sola query RAW contra tablas base.
func (u *UseCases) CheckCostsCoherence(ctx context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error) {
	if filter.ProjectID == nil || *filter.ProjectID <= 0 {
		return nil, domainerr.Validation("project_id is required")
	}
	projectID := *filter.ProjectID

	dashboard, err := u.dashboardRepo.GetDashboard(ctx, dashboardDomain.DashboardFilter{ProjectID: &projectID})
	if err != nil {
		return nil, domainerr.Internal("fetch dashboard: " + err.Error())
	}
	summary := dashboardSummary(dashboard)
	if summary == nil {
		return nil, domainerr.Internal("dashboard summary is empty for project")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	checks := make([]domain.IntegrityCheck, 5)
	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	setErr := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			cancel()
		})
	}

	run := func(index, controlNumber int, fn func(context.Context) (domain.IntegrityCheck, error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			check, err := fn(ctx)
			if err != nil {
				setErr(domainerr.Internal(fmt.Sprintf("control %d failed: %s", controlNumber, err.Error())))
				return
			}
			checks[index] = check
		}()
	}

	run(0, 1, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.checkDirectCosts(c, projectID, summary)
	})
	run(1, 2, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.checkNetIncome(c, projectID, summary)
	})
	run(2, 3, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.checkSupplyInvestment(c, projectID, summary)
	})
	run(3, 4, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.checkAdminCost(c, projectID, summary)
	})
	run(4, 5, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.checkLease(c, projectID, summary)
	})

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	return &domain.IntegrityReport{Checks: checks}, nil
}

func (u *UseCases) GetTentativePrices(ctx context.Context, filter domain.TentativePricesFilter) (*domain.TentativePricesReport, error) {
	items, count, err := u.supplyRepo.ListTentativePrices(ctx, supplyDomain.SupplyFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	}, 10)
	if err != nil {
		return nil, err
	}

	out := make([]domain.TentativePriceItem, len(items))
	for i := range items {
		out[i] = domain.TentativePriceItem{
			SupplyID:     items[i].SupplyID,
			Name:         items[i].Name,
			CategoryName: items[i].CategoryName,
			Price:        items[i].Price,
		}
	}

	return &domain.TentativePricesReport{
		Count: count,
		Items: out,
	}, nil
}

// dashboardSummary extrae el Summary del DashboardData con guardia nil.
func dashboardSummary(d *dashboardDomain.DashboardData) *dashboardDomain.DashboardBalanceSummary {
	if d == nil || d.ManagementBalance == nil {
		return nil
	}
	return d.ManagementBalance.Summary
}

// CONTROL 1: Costos directos ejecutados.
// SSOT: dashboard.DirectCostsExecutedUSD (vía v4_ssot.direct_costs_total_for_project).
// RAW:  ∑(wo.effective_area × labor.price) + ∑(wi.total_used × supply.price) desde tablas base.
func (u *UseCases) checkDirectCosts(ctx context.Context, projectID int64, s *dashboardDomain.DashboardBalanceSummary) (domain.IntegrityCheck, error) {
	raw, err := u.workOrderRepo.GetRawDirectCost(ctx, projectID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	return buildCheck(
		1,
		"Costos directos ejecutados",
		"Dashboard vs cálculo RAW desde órdenes de trabajo",
		"dashboard.DirectCostsExecutedUSD = ∑(wo.effective_area × labor.price + ∑wi.total_used × supply.price)",
		"dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD",
		s.DirectCostsExecutedUSD,
		"v4_report.dashboard_management_balance (v4_ssot.direct_costs_total_for_project)",
		"Costos directos ejecutados que el dashboard muestra en el Cuadro de Gestión.",
		"∑(wo.effective_area × labor.price) + ∑(wi.total_used × supply.price)",
		raw,
		"public.workorders + public.workorder_items + public.labors + public.supplies",
		"Suma directa desde tablas base, respetando deleted_at y tenant_id.",
	), nil
}

// CONTROL 2: Ingreso neto.
// SSOT: dashboard.IncomeUSD (vía v4_ssot.income_net_total_for_lot = lots.tons × net_price).
// RAW:  ∑(lots.tons × crop_commercializations.net_price) directo desde tablas base.
func (u *UseCases) checkNetIncome(ctx context.Context, projectID int64, s *dashboardDomain.DashboardBalanceSummary) (domain.IntegrityCheck, error) {
	raw, err := u.reportRepo.GetRawNetIncome(ctx, projectID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	return buildCheck(
		2,
		"Ingreso neto",
		"Dashboard vs cálculo RAW desde lots + crop_commercializations",
		"dashboard.IncomeUSD = ∑(lots.tons × crop_commercializations.net_price)",
		"dashboard.ManagementBalance.Summary.IncomeUSD",
		s.IncomeUSD,
		"v4_report.dashboard_management_balance (v4_ssot.income_net_total_for_lot)",
		"Ingreso neto total que el dashboard muestra en el Cuadro de Gestión.",
		"∑(lots.tons × crop_commercializations.net_price)",
		raw,
		"public.lots + public.fields + public.crop_commercializations",
		"Suma directa desde tablas base cruzando proyecto y cultivo, respetando deleted_at y tenant_id.",
	), nil
}

// CONTROL 3: Inversión en insumos (semillas + agroquímicos + fertilizantes).
// SSOT: dashboard.SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD.
// RAW:  ∑(supply_movements.quantity × supplies.price) filtrando is_entry y type_id ∈ (1, 2, 3).
func (u *UseCases) checkSupplyInvestment(ctx context.Context, projectID int64, s *dashboardDomain.DashboardBalanceSummary) (domain.IntegrityCheck, error) {
	raw, err := u.supplyRepo.GetRawSupplyInvestment(ctx, projectID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	ssot := s.SemillasInvertidosUSD.Add(s.AgroquimicosInvertidosUSD).Add(s.FertilizantesInvertidosUSD)
	return buildCheck(
		3,
		"Inversión en insumos",
		"Dashboard (semillas + agroquímicos + fertilizantes) vs cálculo RAW desde supply_movements",
		"dashboard.(Sem+Agroq+Fert)InvertidosUSD = ∑(supply_movements.quantity × supplies.price)",
		"dashboard.Summary.(SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD)",
		ssot,
		"v4_report.dashboard_management_balance (v4_ssot.seeds_invested + agrochemicals_invested + fertilizantes_invertidos)",
		"Inversión total en insumos que el dashboard muestra como invertidos en el Cuadro de Gestión.",
		"∑(sm.quantity × s.price) WHERE sm.is_entry AND categories.type_id IN (1,2,3) AND movement_type ∈ whitelist",
		raw,
		"public.supply_movements + public.supplies + public.categories",
		"Suma directa desde tablas base con el mismo filtro de categoría y movement_type que el SSOT.",
	), nil
}

// CONTROL 4: Costo administrativo total del proyecto.
// SSOT: dashboard.StructureExecutedUSD (vía v4_ssot.admin_cost_total_for_project = admin_cost × total_hectares).
// RAW:  projects.admin_cost × Σ(lots.hectares) directo desde tablas base.
func (u *UseCases) checkAdminCost(ctx context.Context, projectID int64, s *dashboardDomain.DashboardBalanceSummary) (domain.IntegrityCheck, error) {
	raw, err := u.projectRepo.GetRawAdminCostTotal(ctx, projectID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	return buildCheck(
		4,
		"Costo administrativo total",
		"Dashboard vs cálculo RAW desde projects.admin_cost × Σhectares",
		"dashboard.StructureExecutedUSD = projects.admin_cost × Σ(lots.hectares)",
		"dashboard.ManagementBalance.Summary.StructureExecutedUSD",
		s.StructureExecutedUSD,
		"v4_report.dashboard_management_balance (v4_ssot.admin_cost_total_for_project)",
		"Costo administrativo total que el dashboard muestra como ejecutado en Estructura.",
		"projects.admin_cost × Σ(lots.hectares)",
		raw,
		"public.projects + public.lots + public.fields",
		"Multiplicación directa del admin_cost del proyecto por la suma de hectáreas de sus lotes activos.",
	), nil
}

// CONTROL 5: Arriendo ejecutado (solo tipos fijos 3 y 4).
// SSOT: dashboard.RentExecutedUSD (vía v4_ssot.lease_executed_for_project).
// RAW:  Σ CASE WHEN fields.lease_type_id IN (3,4) THEN lease_type_value × hectares ELSE 0.
func (u *UseCases) checkLease(ctx context.Context, projectID int64, s *dashboardDomain.DashboardBalanceSummary) (domain.IntegrityCheck, error) {
	raw, err := u.lotRepo.GetRawLeaseExecuted(ctx, projectID)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	return buildCheck(
		5,
		"Arriendo ejecutado",
		"Dashboard vs cálculo RAW desde fields.lease_type fijo (tipos 3 y 4) × lots.hectares",
		"dashboard.RentExecutedUSD = Σ CASE WHEN lease_type_id IN (3,4) THEN lease_type_value × hectares ELSE 0",
		"dashboard.ManagementBalance.Summary.RentExecutedUSD",
		s.RentExecutedUSD,
		"v4_report.dashboard_management_balance (v4_ssot.lease_executed_for_project)",
		"Arriendo ejecutado que el dashboard muestra en el Cuadro de Gestión (solo arriendos fijos).",
		"Σ CASE WHEN fields.lease_type_id IN (3,4) THEN fields.lease_type_value × lots.hectares ELSE 0 END",
		raw,
		"public.lots + public.fields",
		"Suma directa replicando el CASE del SSOT contra las tablas base.",
	), nil
}

// buildCheck construye un IntegrityCheck calculando Status puramente por |dif| > tolerance.
// Sin Status forzado, sin annotateX. SystemValue y RecalcA son las únicas fuentes de verdad.
func buildCheck(
	controlNumber int,
	dataToVerify, description, controlRule string,
	systemCalculation string,
	systemValue decimal.Decimal,
	systemSource string,
	systemMeaning string,
	recalcACalculation string,
	recalcAValue decimal.Decimal,
	recalcASource string,
	recalcAMeaning string,
) domain.IntegrityCheck {
	differenceA := systemValue.Sub(recalcAValue)
	status := checkStatusOK
	severity := checkSeverityInfo
	if differenceA.Abs().GreaterThan(defaultTolerance) {
		status = checkStatusError
		severity = checkSeverityError
	}

	return domain.IntegrityCheck{
		ControlNumber:      controlNumber,
		DataToVerify:       dataToVerify,
		Description:        description,
		ControlRule:        controlRule,
		CheckType:          checkTypeStrong,
		Severity:           severity,
		SystemCalculation:  systemCalculation,
		SystemValue:        systemValue,
		SystemSource:       systemSource,
		SystemMeaning:      systemMeaning,
		RecalcACalculation: recalcACalculation,
		RecalcAValue:       recalcAValue,
		RecalcASource:      recalcASource,
		RecalcAMeaning:     recalcAMeaning,
		DifferenceA:        differenceA,
		Status:             status,
		Tolerance:          defaultTolerance,
	}
}
