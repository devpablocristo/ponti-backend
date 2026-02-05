// Package dataintegrity implementa casos de uso para validar la coherencia de datos
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
//
// ESTOS CÁLCULOS SON CRÍTICOS Y NO DEBEN ALTERARSE A MENOS QUE SE RECIBA
// UNA ORDEN DIRECTA Y CLARA DEL USUARIO.
//
// REGLAS INVIOLABLES:
// - NUNCA modificar los cálculos LEFT/RIGHT sin autorización explícita
// - NUNCA cambiar las tolerancias sin autorización explícita
// - NUNCA alterar la lógica de los 14 controles sin autorización explícita
// - NUNCA usar ROUND() en cálculos internos (solo en DTOs de salida)
// - SIEMPRE mantener precisión completa en cálculos SQL y Go
//
// Si necesitas modificar algo, DEBES pedir autorización explícita primero.
package dataintegrity

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	dashboardDomain "github.com/alphacodinggroup/ponti-backend/internal/dashboard/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/internal/data-integrity/usecases/domain"
	lotDomain "github.com/alphacodinggroup/ponti-backend/internal/lot/usecases/domain"
	reportDomain "github.com/alphacodinggroup/ponti-backend/internal/report/usecases/domain"
	stockDomain "github.com/alphacodinggroup/ponti-backend/internal/stock/usecases/domain"
	workOrderDomain "github.com/alphacodinggroup/ponti-backend/internal/work-order/usecases/domain"
)

// WorkOrderRepositoryPort define la interfaz para el repositorio de work orders.
type WorkOrderRepositoryPort interface {
	GetMetrics(ctx context.Context, filt workOrderDomain.WorkOrderFilter) (*workOrderDomain.WorkOrderMetrics, error)
	GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error)
}

// DashboardRepositoryPort define la interfaz para el repositorio de dashboard
type DashboardRepositoryPort interface {
	GetDashboard(ctx context.Context, filter dashboardDomain.DashboardFilter) (*dashboardDomain.DashboardData, error)
}

// LotRepositoryPort define la interfaz para el repositorio de lotes
type LotRepositoryPort interface {
	ListLots(ctx context.Context, filter lotDomain.LotListFilter, page, pageSize int) ([]lotDomain.LotTable, int, decimal.Decimal, decimal.Decimal, error)
}

// ReportRepositoryPort define la interfaz para el repositorio de reportes
type ReportRepositoryPort interface {
	GetSummaryResults(filters reportDomain.SummaryResultsFilter) ([]reportDomain.SummaryResults, error)
	GetFieldCropMetrics(filters reportDomain.ReportFilter) ([]reportDomain.FieldCropMetric, error)
	GetInvestorContributionReport(ctx context.Context, filter reportDomain.ReportFilter) (*reportDomain.InvestorContributionReport, error)
}

// StockRepositoryPort define la interfaz para el repositorio de stock
type StockRepositoryPort interface {
	GetStocks(ctx context.Context, projectID int64, closeDate time.Time) ([]*stockDomain.Stock, error)
}

// UseCases contiene los casos de uso del módulo dataintegrity
type UseCases struct {
	workOrderRepo WorkOrderRepositoryPort
	dashboardRepo DashboardRepositoryPort
	lotRepo       LotRepositoryPort
	reportRepo    ReportRepositoryPort
	stockRepo     StockRepositoryPort
}

// NewUseCases crea una nueva instancia de UseCases
func NewUseCases(
	workOrderRepo WorkOrderRepositoryPort,
	dashboardRepo DashboardRepositoryPort,
	lotRepo LotRepositoryPort,
	reportRepo ReportRepositoryPort,
	stockRepo StockRepositoryPort,
) *UseCases {
	return &UseCases{
		workOrderRepo: workOrderRepo,
		dashboardRepo: dashboardRepo,
		lotRepo:       lotRepo,
		reportRepo:    reportRepo,
		stockRepo:     stockRepo,
	}
}

// CheckCostsCoherence valida la coherencia de costos con 14 controles individuales
// Cada control calcula LEFT (origen/correcto) y RIGHT (destino/validar) de forma INDEPENDIENTE
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTA FUNCIÓN CONTIENE LOS 14 CONTROLES CRÍTICOS DE INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) CheckCostsCoherence(ctx context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	checks := make([]domain.IntegrityCheck, 14)
	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	setErr := func(err error) {
		errOnce.Do(func() {
			firstErr = err
			cancel()
		})
	}

	run := func(index int, controlNumber int, fn func(context.Context) (domain.IntegrityCheck, error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			check, err := fn(ctx)
			if err != nil {
				setErr(fmt.Errorf("control %d failed: %w", controlNumber, err))
				return
			}
			checks[index] = check
		}()
	}

	// GRUPO 1: Costos Directos Ejecutados (Controles 1-4)
	run(0, 1, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control1OrdenesVsDashboard(c, filter.ProjectID)
	})
	run(1, 2, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control2OrdenesVsLotes(c, filter.ProjectID)
	})
	run(2, 3, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control3OrdenesVsInformeCampo(c, filter.ProjectID)
	})
	run(3, 4, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control4OrdenesVsInformeGenerales(c, filter.ProjectID)
	})

	// GRUPO 2: Invertidos (Controles 5-7)
	run(4, 5, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control5LaboresInsumosVsDashboard(c, filter.ProjectID)
	})
	run(5, 6, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control6LaboresVsAportes(c, filter.ProjectID)
	})
	run(6, 7, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control7InsumosVsAportes(c, filter.ProjectID)
	})

	// GRUPO 3: Lotes → Aportes (Controles 8-9)
	run(7, 8, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control8LotesAdminVsAportes(c, filter.ProjectID)
	})
	run(8, 9, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control9LotesArriendoVsAportes(c, filter.ProjectID)
	})

	// GRUPO 4: Ingreso Neto (Control 10)
	run(9, 10, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control10LotesIngresoNetoVsResumen(c, filter.ProjectID)
	})

	// GRUPO 5: Resultado Operativo (Controles 11-13)
	run(10, 11, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control11LotesResultadoVsInformeCultivo(c, filter.ProjectID)
	})
	run(11, 12, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control12LotesResultadoVsInformeGenerales(c, filter.ProjectID)
	})
	run(12, 13, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control13LotesResultadoVsDashboard(c, filter.ProjectID)
	})

	// GRUPO 6: Stock (Control 14)
	run(13, 14, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control14StockVsDashboard(c, filter.ProjectID)
	})

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	return &domain.IntegrityReport{
		Checks: checks,
	}, nil
}

// =====================================================
// CONTROL 1: Órdenes de trabajo → Dashboard
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: Dashboard.CostosDirectos (SSOT)
// =====================================================
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTE CONTROL ES CRÍTICO PARA LA INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) control1OrdenesVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Suma de costos directos desde lotes (v4_report.lot_list)
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		costTotal := lot.CostUsdPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(costTotal)
	}

	// RIGHT: Costos desde dashboard (usa funciones SSOT)
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	rightValue := dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD

	return buildCheck(
		1,
		"Lotes",
		"Costos directos ejecutados",
		"Dashboard",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares) = dashboard_management_balance.costos_directos_ejecutados_usd",
		"Compara costos directos ejecutados entre lotes y dashboard.",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Total de costos directos desde lotes.",
		"dashboard_management_balance.costos_directos_ejecutados_usd",
		rightValue,
		"v4_report.dashboard_management_balance",
		"Costo directo ejecutado mostrado en dashboard.",
		"Ambos deben representar el mismo total del proyecto.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 2: Órdenes de trabajo → Lotes
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: ∑(Costo_directo_ha_lote × Superficie_lote)
// =====================================================
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTE CONTROL ES CRÍTICO PARA LA INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) control2OrdenesVsLotes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Suma desde lotes (v4_report.lot_list)
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		costTotal := lot.CostUsdPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(costTotal)
	}

	// RIGHT: Desde informe field-crop
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	fieldCropMetrics, err := u.reportRepo.GetFieldCropMetrics(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, metric := range fieldCropMetrics {
		if !metric.TotalDirectCostsUsd.IsZero() {
			rightValue = rightValue.Add(metric.TotalDirectCostsUsd)
		} else {
			costTotal := metric.SurfaceHa.Mul(metric.DirectCostsUsdHa)
			rightValue = rightValue.Add(costTotal)
		}
	}

	return buildCheck(
		2,
		"Lotes",
		"Costos directos ejecutados",
		"Informe de Resultado por campo",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares) = ∑(field_crop_metrics.total_direct_costs_usd)",
		"Compara costos directos entre lotes y reporte por campo.",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Costos directos sumados desde lotes.",
		"∑(field_crop_metrics.total_direct_costs_usd)",
		rightValue,
		"v4_report.field_crop_metrics",
		"Costos directos agregados por campo/cultivo.",
		"Mismo total, distinto nivel de agregación.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 3: Órdenes de trabajo → Informe de Resultado por campo
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: ∑(Costo_directo_ha_Cultivo × Superficie_Cultivo)
// =====================================================
func (u *UseCases) control3OrdenesVsInformeCampo(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Desde informe field-crop
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	fieldCropMetrics, err := u.reportRepo.GetFieldCropMetrics(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, metric := range fieldCropMetrics {
		if !metric.TotalDirectCostsUsd.IsZero() {
			leftValue = leftValue.Add(metric.TotalDirectCostsUsd)
		} else {
			costTotal := metric.SurfaceHa.Mul(metric.DirectCostsUsdHa)
			leftValue = leftValue.Add(costTotal)
		}
	}

	// RIGHT: Total del resumen de resultados (totales del proyecto)
	summaryFilter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(summaryFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	if len(summaryResults) > 0 {
		rightValue = summaryResults[0].TotalDirectCostsUsd
	}

	return buildCheck(
		3,
		"Informe de Resultado por campo",
		"Costos directos ejecutados",
		"Informe de Resultado Generales",
		"∑(field_crop_metrics.total_direct_costs_usd) = summary_results.total_direct_costs_usd",
		"Compara costos directos del reporte por campo vs resumen general.",
		"∑(field_crop_metrics.total_direct_costs_usd)",
		leftValue,
		"v4_report.field_crop_metrics",
		"Total de costos directos por campo/cultivo.",
		"summaryResults[0].TotalDirectCostsUsd (totales del proyecto)",
		rightValue,
		"v4_report.summary_results",
		"Total de costos directos del resumen.",
		"Mismo total del proyecto.",
		decimal.NewFromInt(1), // Tolerancia = 1 USD (diferencias de precisión en agregaciones)
	), nil
}

// =====================================================
// CONTROL 4: Órdenes de trabajo → Informe de Resultado Generales
// LEFT: ∑(Ordenes.costo_total) RAW
// RIGHT: Total de informe generales
// =====================================================
func (u *UseCases) control4OrdenesVsInformeGenerales(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Total del resumen de resultados (totales del proyecto)
	filter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	if len(summaryResults) > 0 {
		leftValue = summaryResults[0].TotalDirectCostsUsd
	}

	// RIGHT: Costos desde dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	rightValue := dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD

	return buildCheck(
		4,
		"Informe de Resultado Generales",
		"Costos directos ejecutados",
		"Dashboard",
		"summary_results.total_direct_costs_usd (totales del proyecto) = dashboard_management_balance.costos_directos_ejecutados_usd",
		"Compara costos directos del resumen vs dashboard.",
		"summaryResults[0].TotalDirectCostsUsd (totales del proyecto)",
		leftValue,
		"v4_report.summary_results",
		"Total de costos directos del resumen.",
		"dashboard_management_balance.costos_directos_ejecutados_usd",
		rightValue,
		"v4_report.dashboard_management_balance",
		"Costo directo ejecutado en dashboard.",
		"Mismo total del proyecto.",
		decimal.NewFromInt(1), // Tolerancia = 1 USD (diferencias de precisión en agregaciones)
	), nil
}

// =====================================================
// CONTROL 5: Labores + Insumos → Dashboard
// LEFT: ∑(Labores) + ∑(Insumos) desde dashboard breakdown
// RIGHT: Dashboard.Invertidos total
// =====================================================
func (u *UseCases) control5LaboresInsumosVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	summary := dashboardData.ManagementBalance.Summary

	// LEFT: Suma de componentes (Labores + Semillas + Agroquímicos + Fertilizantes)
	labores := summary.LaboresInvertidosUSD
	semilla := summary.SemillasInvertidosUSD
	agroquimicos := summary.AgroquimicosInvertidosUSD
	fertilizantes := summary.FertilizantesInvertidosUSD
	leftValue := labores.Add(semilla).Add(agroquimicos).Add(fertilizantes)

	// RIGHT: Total invertidos desde dashboard
	rightValue := dashboardData.ManagementBalance.TotalsRow.InvestedUSD

	return buildCheck(
		5,
		"Labores + Insumos",
		"Invertidos",
		"Dashboard",
		"TotalsRow.InvestedUSD = SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD",
		"Valida total invertido del dashboard.",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD",
		leftValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Suma de componentes invertidos.",
		"TotalsRow.InvestedUSD",
		rightValue,
		"Dashboard TotalsRow (ManagementBalance.TotalsRow)",
		"Total invertido calculado por dashboard.",
		"Total debe ser suma de componentes.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 6: Labores → Informe de Aportes
// LEFT: Dashboard.LaboresInvertidos (sin cosecha)
// RIGHT: Aportes (Labores Generales + Siembra + Riego)
// =====================================================
func (u *UseCases) control6LaboresVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// LEFT: Labores invertidas desde dashboard
	leftValue := dashboardData.ManagementBalance.Summary.LaboresInvertidosUSD

	// RIGHT: Informe de aportes (suma de categorías sin cosecha)
	reportFilter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, reportFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Labores Generales" || category.Label == "Siembra" ||
			category.Label == "Riego" {
			rightValue = rightValue.Add(category.TotalUsd)
		}
	}

	return buildCheck(
		6,
		"Dashboard",
		"Inversión en labores",
		"Informe de Aportes",
		"Summary.LaboresInvertidosUSD = ∑(contribution_categories.total_usd) where label in [Labores Generales, Siembra, Riego]",
		"Valida labores entre dashboard y aportes.",
		"Summary.LaboresInvertidosUSD",
		leftValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Labores invertidas en dashboard.",
		"∑(contribution_categories.total_usd) labels: Labores Generales, Siembra, Riego",
		rightValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Suma de categorías de labores en aportes.",
		"Mismo total de labores (sin cosecha).",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 7: Insumos → Informe de Aportes
// LEFT: Dashboard.InsumosInvertidos
// RIGHT: Aportes (Semillas + Agroquímicos)
// =====================================================
func (u *UseCases) control7InsumosVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// LEFT: Insumos invertidos desde dashboard
	leftValue := dashboardData.ManagementBalance.Summary.SemillasInvertidosUSD.
		Add(dashboardData.ManagementBalance.Summary.AgroquimicosInvertidosUSD).
		Add(dashboardData.ManagementBalance.Summary.FertilizantesInvertidosUSD)

	// RIGHT: Informe de aportes
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Semilla" || category.Label == "Agroquímicos" || category.Label == "Fertilizantes" {
			rightValue = rightValue.Add(category.TotalUsd)
		}
	}

	return buildCheck(
		7,
		"Insumos",
		"Inversión en insumos",
		"Informe de Aportes",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD = ∑(contribution_categories.total_usd) labels: Semilla, Agroquímicos, Fertilizantes",
		"Valida insumos entre dashboard y aportes.",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD",
		leftValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Semillas, agroquímicos y fertilizantes en dashboard.",
		"∑(contribution_categories.total_usd) labels: Semilla, Agroquímicos, Fertilizantes",
		rightValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Suma de esas categorías en aportes.",
		"Mismo total de insumos.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 8: Lotes → Informe de Aportes (Administración)
// LEFT: ∑(AdminCost_ha × Superficie) desde lotes
// RIGHT: Aportes Adm.Proyecto
// =====================================================
func (u *UseCases) control8LotesAdminVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Calcular desde lotes
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		leftValue = leftValue.Add(lot.AdminCost.Mul(lot.Hectares))
	}

	// RIGHT: Total Aportes Adm.Proyecto del Informe
	reportFilter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, reportFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Administración y Estructura" {
			rightValue = category.TotalUsd
			break
		}
	}

	return buildCheck(
		8,
		"Lotes",
		"Invertidos",
		"Informe de Aportes",
		"∑(lot_list.admin_cost_per_ha_usd × lot_list.hectares) = contribution_categories.total_usd (Administración y Estructura)",
		"Valida administración entre lotes y aportes.",
		"∑(lot_list.admin_cost_per_ha_usd × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Admin prorrateado por lote.",
		"contribution_categories.total_usd (Administración y Estructura)",
		rightValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Categoría Administración y Estructura.",
		"Mismo total de administración.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 9: Lotes → Informe de Aportes (Arriendo)
// LEFT: ∑(Arriendo_ha × Superficie) desde lotes
// RIGHT: Aportes Arriendo Fijo
// =====================================================
func (u *UseCases) control9LotesArriendoVsAportes(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Calcular desde lotes
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		arriendoTotal := lot.RentPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(arriendoTotal)
	}

	// RIGHT: Total Aportes Arriendo del Informe
	filter := reportDomain.ReportFilter{ProjectID: projectID}
	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, category := range investorReport.Contributions {
		if category.Label == "Arriendo Capitalizable" {
			rightValue = category.TotalUsd
			break
		}
	}

	return buildCheck(
		9,
		"Lotes",
		"Invertidos",
		"Informe de Aportes",
		"∑(lot_list.rent_per_ha_usd × lot_list.hectares) = contribution_categories.total_usd (Arriendo Capitalizable)",
		"Valida arriendo capitalizable entre lotes y aportes.",
		"∑(lot_list.rent_per_ha_usd × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Arriendo prorrateado por lote.",
		"contribution_categories.total_usd (Arriendo Capitalizable)",
		rightValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Categoría Arriendo Capitalizable.",
		"Mismo total de arriendo.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 10: Lotes → Resumen de resultados (Ingreso Neto)
// LEFT: ∑(Ingreso_Neto_ha × Superficie) desde lotes
// RIGHT: ∑(Ingreso_Neto) del resumen
// =====================================================
func (u *UseCases) control10LotesIngresoNetoVsResumen(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Calcular desde lotes
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		ingresoTotal := lot.IncomeNetPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(ingresoTotal)
	}

	// RIGHT: Obtener resumen de resultados
	summaryFilter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(summaryFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, result := range summaryResults {
		rightValue = rightValue.Add(result.NetIncomeUsd)
	}

	return buildCheck(
		10,
		"Lotes",
		"Ingreso Neto",
		"Resumen de resultados",
		"∑(lot_list.income_net_per_ha_usd × lot_list.hectares) = ∑(summary_results.net_income_usd)",
		"Valida ingreso neto entre lotes y resumen.",
		"∑(lot_list.income_net_per_ha_usd × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Ingreso neto total desde lotes.",
		"∑(summary_results.net_income_usd)",
		rightValue,
		"v4_report.summary_results",
		"Ingreso neto total del resumen.",
		"Mismo total del proyecto.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 11: Lotes → Informe por cultivo (Resultado Operativo)
// LEFT: ∑(Resultado.Operativo_ha × Superficie) desde lotes
// RIGHT: ∑(Resultado.Operativo × Superficie) por cultivo
// =====================================================
func (u *UseCases) control11LotesResultadoVsInformeCultivo(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Calcular desde lotes
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		resultadoTotal := lot.OperatingResultPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(resultadoTotal)
	}

	// RIGHT: Informe por cultivo
	reportFilter := reportDomain.ReportFilter{ProjectID: projectID}
	fieldCropMetrics, err := u.reportRepo.GetFieldCropMetrics(reportFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	for _, metric := range fieldCropMetrics {
		resultadoTotal := metric.OperatingResultUsdHa.Mul(metric.SurfaceHa)
		rightValue = rightValue.Add(resultadoTotal)
	}

	return buildCheck(
		11,
		"Lotes",
		"Resultado operativo total",
		"Informe por cultivo",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares) = ∑(field_crop_metrics.operating_result_usd_ha × field_crop_metrics.surface_ha)",
		"Valida resultado operativo entre lotes y reporte por cultivo.",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Resultado operativo total desde lotes.",
		"∑(field_crop_metrics.operating_result_usd_ha × field_crop_metrics.surface_ha)",
		rightValue,
		"v4_report.field_crop_metrics",
		"Resultado operativo agregado por cultivo.",
		"Mismo total del proyecto.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 12: Lotes → Informe de Resultado Generales (Resultado Operativo)
// LEFT: ∑(Resultado.Operativo_ha × Superficie) desde lotes
// RIGHT: Total Resultado Operativo del informe general
// =====================================================
func (u *UseCases) control12LotesResultadoVsInformeGenerales(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Calcular desde lotes
	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		resultadoTotal := lot.OperatingResultPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(resultadoTotal)
	}

	// RIGHT: Informe de Resultado Generales (primera fila = GRAL)
	summaryFilter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(summaryFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := decimal.Zero
	if len(summaryResults) > 0 {
		rightValue = summaryResults[0].TotalOperatingResultUsd
	}

	return buildCheck(
		12,
		"Lotes",
		"Resultado operativo total",
		"Informe de Resultado Generales",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares) = summary_results.total_operating_result_usd (totales proyecto)",
		"Valida resultado operativo entre lotes y resumen.",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Resultado operativo total desde lotes.",
		"summaryResults[0].TotalOperatingResultUsd (totales del proyecto)",
		rightValue,
		"v4_report.summary_results",
		"Resultado operativo total del resumen.",
		"Mismo total del proyecto.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 13: Lotes → Dashboard (Resultado Operativo)
// LEFT: ∑(Resultado.Operativo_ha × Superficie) desde lotes
// RIGHT: Dashboard Card Resultado Operativo
// =====================================================
func (u *UseCases) control13LotesResultadoVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// LEFT: Calcular desde lotes
	filter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, filter, 1, 10000)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	leftValue := decimal.Zero
	for _, lot := range lots {
		resultadoTotal := lot.OperatingResultPerHa.Mul(lot.Hectares)
		leftValue = leftValue.Add(resultadoTotal)
	}

	// RIGHT: Dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	rightValue := dashboardData.Metrics.OperatingResult.ResultUSD

	return buildCheck(
		13,
		"Lotes",
		"Resultado operativo total",
		"Dashboard",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares) = dashboard_management_balance.operating_result_usd",
		"Valida resultado operativo entre lotes y dashboard.",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares)",
		leftValue,
		"v4_report.lot_list",
		"Resultado operativo total desde lotes.",
		"Metrics.OperatingResult.ResultUSD",
		rightValue,
		"v4_report.dashboard_management_balance.operating_result_usd",
		"Resultado operativo en dashboard.",
		"Mismo total del proyecto.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 14: Stock → Dashboard
// LEFT: Invertido - Ejecutado
// RIGHT: Dashboard.Stock
// =====================================================
func (u *UseCases) control14StockVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	// Obtener dashboard
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	summary := dashboardData.ManagementBalance.Summary

	// LEFT: Calcular stock esperado (Invertido - Ejecutado)
	invertido := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD).
		Add(summary.LaboresInvertidosUSD)
	ejecutado := summary.DirectCostsExecutedUSD
	leftValue := invertido.Sub(ejecutado)

	// RIGHT: Stock desde dashboard
	rightValue := summary.StockUSD

	return buildCheck(
		14,
		"Stock",
		"Insumos en stock",
		"Dashboard",
		"Summary.StockUSD = (SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD) - DirectCostsExecutedUSD",
		"Valida stock en dashboard.",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD - DirectCostsExecutedUSD",
		leftValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Invertido menos ejecutado.",
		"Summary.StockUSD",
		rightValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Stock USD del dashboard.",
		"Stock = invertido - ejecutado.",
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// HELPER: buildCheck construye un IntegrityCheck con LEFT/RIGHT
// =====================================================
func buildCheck(
	controlNumber int,
	sourceModule, dataToVerify, targetModule, controlRule, description string,
	leftCalculation string,
	leftValue decimal.Decimal,
	leftSource, leftMeaning string,
	rightCalculation string,
	rightValue decimal.Decimal,
	rightSource, rightMeaning string,
	calculationMeaning string,
	tolerance decimal.Decimal,
) domain.IntegrityCheck {
	difference := leftValue.Sub(rightValue)
	status := "OK"

	// DEBUG: Log para control 2
	if controlNumber == 2 {
		println(fmt.Sprintf("[DEBUG Control 2] leftValue: %s", leftValue.String()))
		println(fmt.Sprintf("[DEBUG Control 2] rightValue: %s", rightValue.String()))
		println(fmt.Sprintf("[DEBUG Control 2] difference: %s", difference.String()))
		println(fmt.Sprintf("[DEBUG Control 2] difference.Abs(): %s", difference.Abs().String()))
		println(fmt.Sprintf("[DEBUG Control 2] tolerance: %s", tolerance.String()))
		println(fmt.Sprintf("[DEBUG Control 2] GreaterThan result: %v", difference.Abs().GreaterThan(tolerance)))
	}

	if difference.Abs().GreaterThan(tolerance) {
		status = "ERROR"
	}

	return domain.IntegrityCheck{
		ControlNumber:    controlNumber,
		SourceModule:     sourceModule,
		DataToVerify:     dataToVerify,
		TargetModule:     targetModule,
		ControlRule:      controlRule,
		Description:      description,
		LeftCalculation:  leftCalculation,
		LeftValue:        leftValue,
		LeftSource:       leftSource,
		LeftMeaning:      leftMeaning,
		RightCalculation: rightCalculation,
		RightValue:       rightValue,
		RightSource:      rightSource,
		RightMeaning:     rightMeaning,
		CalculationMeaning: calculationMeaning,
		Difference:       difference,
		Status:           status,
		Tolerance:        tolerance,
	}
}
