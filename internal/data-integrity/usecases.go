// Package dataintegrity implementa controles de coherencia entre módulos (integridad de datos).
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

// sharedData cachea datos compartidos entre controles para reducir round-trips a DB.
type sharedData struct {
	lots             []lotDomain.LotTable
	dashboardData    *dashboardDomain.DashboardData
	fieldCropMetrics []reportDomain.FieldCropMetric
	summaryResults   []reportDomain.SummaryResults
	investorReport   *reportDomain.InvestorContributionReport
	workOrderMetrics *workOrderDomain.WorkOrderMetrics
	workOrderRawCost decimal.Decimal
}

// fetchSharedData obtiene una sola vez los datos usados por múltiples controles.
func (u *UseCases) fetchSharedData(ctx context.Context, projectID *int64) (*sharedData, error) {
	sd := &sharedData{}

	lotFilter := lotDomain.LotListFilter{ProjectID: projectID}
	lots, _, _, _, err := u.lotRepo.ListLots(ctx, lotFilter, 1, 10000)
	if err != nil {
		return nil, fmt.Errorf("fetch lots: %w", err)
	}
	sd.lots = lots

	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return nil, fmt.Errorf("fetch dashboard: %w", err)
	}
	sd.dashboardData = dashboardData

	reportFilter := reportDomain.ReportFilter{ProjectID: projectID}
	fieldCropMetrics, err := u.reportRepo.GetFieldCropMetrics(reportFilter)
	if err != nil {
		return nil, fmt.Errorf("fetch field_crop_metrics: %w", err)
	}
	sd.fieldCropMetrics = fieldCropMetrics

	summaryFilter := reportDomain.SummaryResultsFilter{ProjectID: projectID}
	summaryResults, err := u.reportRepo.GetSummaryResults(summaryFilter)
	if err != nil {
		return nil, fmt.Errorf("fetch summary_results: %w", err)
	}
	sd.summaryResults = summaryResults

	investorReport, err := u.reportRepo.GetInvestorContributionReport(ctx, reportFilter)
	if err != nil {
		return nil, fmt.Errorf("fetch investor_contribution_report: %w", err)
	}
	sd.investorReport = investorReport

	// Work orders: métricas (v4_report.workorder_metrics) + cálculo RAW independiente.
	woFilter := workOrderDomain.WorkOrderFilter{ProjectID: projectID}
	woMetrics, err := u.workOrderRepo.GetMetrics(ctx, woFilter)
	if err != nil {
		return nil, fmt.Errorf("fetch workorder_metrics: %w", err)
	}
	sd.workOrderMetrics = woMetrics

	rawProjectID := int64(0)
	if projectID != nil {
		rawProjectID = *projectID
	}
	rawCost, err := u.workOrderRepo.GetRawDirectCost(ctx, rawProjectID)
	if err != nil {
		return nil, fmt.Errorf("fetch workorder_raw_direct_cost: %w", err)
	}
	sd.workOrderRawCost = rawCost

	return sd, nil
}

// CheckCostsCoherence valida la coherencia de costos con 16 controles individuales.
// Cada control compara: SystemValue (1 directo) vs RecalcA y opcionalmente RecalcB (2 independientes).
func (u *UseCases) CheckCostsCoherence(ctx context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sd, err := u.fetchSharedData(ctx, filter.ProjectID)
	if err != nil {
		return nil, err
	}

	checks := make([]domain.IntegrityCheck, 16)
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
	// 4 fuentes: lot_list, field_crop_metrics, summary_results, dashboard
	run(0, 1, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control1LotesVsDashboard(c, sd)
	})
	run(1, 2, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control2LotesVsInformeCampo(c, sd)
	})
	run(2, 3, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control3InformeCampoVsInformeGenerales(c, sd)
	})
	run(3, 4, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control4InformeGeneralesVsDashboard(c, sd)
	})

	// GRUPO 2: Invertidos (Controles 5-7)
	run(4, 5, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control5LaboresInsumosVsDashboard(c, sd)
	})
	run(5, 6, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control6LaboresVsAportes(c, sd)
	})
	run(6, 7, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control7InsumosVsAportes(c, sd)
	})

	// GRUPO 3: Lotes → Aportes (Controles 8-9)
	run(7, 8, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control8LotesAdminVsAportes(c, sd)
	})
	run(8, 9, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control9LotesArriendoVsAportes(c, sd)
	})

	// GRUPO 4: Ingreso Neto (Control 10)
	run(9, 10, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control10LotesIngresoNetoVsResumen(c, sd)
	})

	// GRUPO 5: Resultado Operativo (Controles 11-13)
	// 4 fuentes: lot_list, field_crop_metrics, summary_results, dashboard
	run(10, 11, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control11LotesResultadoVsInformeCultivo(c, sd)
	})
	run(11, 12, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control12LotesResultadoVsInformeGenerales(c, sd)
	})
	run(12, 13, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control13LotesResultadoVsDashboard(c, sd)
	})

	// GRUPO 6: Stock (Control 14)
	run(13, 14, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control14StockVsDashboard(c, sd)
	})

	// GRUPO 7: Órdenes de trabajo (Control 15)
	run(14, 15, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control15DashboardVsWorkOrdersDirectCosts(c, sd)
	})

	// GRUPO 8: Costos directos aportados (Control 16)
	run(15, 16, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control16DashboardAportadoVsLaborsAndSupplies(c, sd)
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
// Helpers para calcular valores compartidos
// =====================================================

// sumLotsCostDirecto calcula ∑(lot.cost_usd_per_ha × hectares)
func sumLotsCostDirecto(lots []lotDomain.LotTable) decimal.Decimal {
	total := decimal.Zero
	for _, lot := range lots {
		total = total.Add(lot.CostUsdPerHa.Mul(lot.Hectares))
	}
	return total
}

// sumFieldCropDirectCosts calcula ∑(field_crop_metrics.total_direct_costs_usd)
func sumFieldCropDirectCosts(metrics []reportDomain.FieldCropMetric) decimal.Decimal {
	total := decimal.Zero
	for _, m := range metrics {
		if !m.TotalDirectCostsUsd.IsZero() {
			total = total.Add(m.TotalDirectCostsUsd)
		} else {
			total = total.Add(m.SurfaceHa.Mul(m.DirectCostsUsdHa))
		}
	}
	return total
}

// sumLotsOperatingResult calcula ∑(lot.operating_result_per_ha × hectares)
func sumLotsOperatingResult(lots []lotDomain.LotTable) decimal.Decimal {
	total := decimal.Zero
	for _, lot := range lots {
		total = total.Add(lot.OperatingResultPerHa.Mul(lot.Hectares))
	}
	return total
}

// sumFieldCropOperatingResult calcula ∑(field_crop_metrics.operating_result_usd_ha × surface_ha)
func sumFieldCropOperatingResult(metrics []reportDomain.FieldCropMetric) decimal.Decimal {
	total := decimal.Zero
	for _, m := range metrics {
		total = total.Add(m.OperatingResultUsdHa.Mul(m.SurfaceHa))
	}
	return total
}

// =====================================================
// CONTROL 15: Dashboard (ejecutados) → WorkOrders (costos directos)
// System: dashboard.DirectCostsExecutedUSD
// RecalcA: ∑(workorder_metrics.direct_cost_usd)
// RecalcB: RAW desde tablas workorders + workorder_items
// =====================================================
func (u *UseCases) control15DashboardVsWorkOrdersDirectCosts(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sd.dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD

	recalcA := decimal.Zero
	if sd.workOrderMetrics != nil {
		recalcA = sd.workOrderMetrics.DirectCost
	}
	recalcBVal := sd.workOrderRawCost

	return buildCheck(
		15,
		"Costos directos ejecutados",
		"Dashboard vs Órdenes de trabajo (métricas + RAW)",
		"dashboard.DirectCostsExecutedUSD = ∑(workorder_metrics.direct_cost_usd) = ∑(wo RAW cost)",
		"dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD",
		systemValue,
		"v4_report.dashboard_management_balance",
		"Costos directos ejecutados del Dashboard (SSOT).",
		"∑(workorder_metrics.direct_cost_usd)",
		recalcA,
		"v4_report.workorder_metrics",
		"Recálculo: suma de costos directos por lote/campo desde la vista de métricas de órdenes de trabajo.",
		strPtr("∑(wo.effective_area*l.price + ∑(wi.total_used*s.price))"),
		&recalcBVal,
		strPtr("public.workorders + public.workorder_items + public.supplies + public.labors"),
		strPtr("Segundo recálculo (RAW): calcula el costo directo desde tablas base (labor + insumos), respetando deleted_at y sin depender de vistas SSOT."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 16: Dashboard (aportado) → Labores + Insumos (neto) + Aportes (3ra vía)
// System: dashboard.DirectCostsInvestedUSD
// RecalcA: dashboard.(LaboresInvertidosUSD + SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD)
// RecalcB: ∑(investor_contribution categories) para labores + insumos
// =====================================================
func (u *UseCases) control16DashboardAportadoVsLaborsAndSupplies(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	summary := sd.dashboardData.ManagementBalance.Summary

	systemValue := summary.DirectCostsInvestedUSD

	// RecalcA: sumar "Total USD / Neto" de Labores + Insumos tal como el dashboard los expone.
	// Insumos neto = semillas + agroquímicos + fertilizantes (en invertidos).
	insumosNeto := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD)
	laboresNeto := summary.LaboresInvertidosUSD
	recalcA := laboresNeto.Add(insumosNeto)

	// RecalcB (3ra vía): sumar categorías desde el reporte de aportes (independiente del dashboard).
	var recalcBCalc, recalcBSrc, recalcBMeaning *string
	var recalcBVal *decimal.Decimal
	if sd.investorReport != nil {
		total := decimal.Zero
		for _, cat := range sd.investorReport.Contributions {
			// Labores (mismo set usado en control 6)
			if cat.Label == "Labores Generales" || cat.Label == "Siembra" || cat.Label == "Riego" {
				total = total.Add(cat.TotalUsd)
				continue
			}
			// Insumos
			if cat.Label == "Semilla" || cat.Label == "Agroquímicos" || cat.Label == "Fertilizantes" {
				total = total.Add(cat.TotalUsd)
				continue
			}
		}
		recalcBCalc = strPtr("∑(contribution_categories.total_usd) labels: Labores + Insumos")
		recalcBVal = &total
		recalcBSrc = strPtr("v4_report.investor_contribution_data.contribution_categories")
		recalcBMeaning = strPtr("Tercera vía: suma aportes por categorías (Labores + Insumos) desde el reporte de aportes. Camino independiente al dashboard.")
	}

	return buildCheck(
		16,
		"Costos directos aportados",
		"Dashboard aportado vs Labores+Insumos (neto) vs Aportes",
		"dashboard.DirectCostsInvestedUSD = (LaboresNeto + InsumosNeto) = ∑(aportes: labores + insumos)",
		"dashboard.ManagementBalance.Summary.DirectCostsInvestedUSD",
		systemValue,
		"v4_report.dashboard_management_balance",
		"Costos directos aportados/invertidos del Dashboard.",
		"dashboard.(LaboresInvertidosUSD + SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD)",
		recalcA,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Recálculo: total neto de labores + insumos (semillas+agroquímicos+fertilizantes) desde el resumen del dashboard.",
		recalcBCalc, recalcBVal, recalcBSrc, recalcBMeaning,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 1: Lotes → Dashboard (Costos Directos)
// System: dashboard.DirectCostsExecutedUSD
// RecalcA: ∑(lot.cost×ha)
// RecalcB: summary_results.TotalDirectCostsUsd
// =====================================================
func (u *UseCases) control1LotesVsDashboard(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sd.dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD
	recalcA := sumLotsCostDirecto(sd.lots)

	var recalcBCalc, recalcBSrc, recalcBMeaning *string
	var recalcBVal *decimal.Decimal
	if len(sd.summaryResults) > 0 {
		v := sd.summaryResults[0].TotalDirectCostsUsd
		recalcBCalc = strPtr("summary_results[0].TotalDirectCostsUsd")
		recalcBVal = &v
		recalcBSrc = strPtr("v4_report.summary_results")
		recalcBMeaning = strPtr("Segundo recálculo: total de costos directos del resumen general (summary_results), agrega por cultivo desde field_crop_metrics.")
	}

	return buildCheck(
		1,
		"Costos directos ejecutados",
		"Dashboard vs Lotes vs Informe Generales",
		"dashboard.DirectCostsExecutedUSD = ∑(lot.cost×ha) = summary_results.TotalDirectCostsUsd",
		"dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD",
		systemValue,
		"v4_report.dashboard_management_balance",
		"Costos directos ejecutados del Cuadro de Gestión del Dashboard. Se calcula sumando labores e insumos ejecutados por lote usando funciones SSOT.",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo: costo por hectárea de cada lote × hectáreas, camino workorder_metrics_raw → lot_base_costs → lot_metrics → lot_list.",
		recalcBCalc, recalcBVal, recalcBSrc, recalcBMeaning,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 2: Lotes → Informe por campo (Costos Directos)
// System: ∑(lot.cost×ha)
// RecalcA: ∑(field_crop_metrics.total_direct_costs)
// RecalcB: dashboard.DirectCostsExecutedUSD
// =====================================================
func (u *UseCases) control2LotesVsInformeCampo(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sumLotsCostDirecto(sd.lots)
	recalcA := sumFieldCropDirectCosts(sd.fieldCropMetrics)
	recalcBVal := sd.dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD

	return buildCheck(
		2,
		"Costos directos ejecutados",
		"Lotes vs Informe por campo vs Dashboard",
		"∑(lot.cost×ha) = ∑(field_crop.total_direct_costs) = dashboard.DirectCostsExecutedUSD",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares)",
		systemValue,
		"v4_report.lot_list",
		"Suma total de costos directos desde tabla de lotes. Cada lote tiene costo/ha calculado en lot_metrics.",
		"∑(field_crop_metrics.total_direct_costs_usd)",
		recalcA,
		"v4_report.field_crop_metrics",
		"Recálculo: costos directos agregados por campo/cultivo desde field_crop_aggregated → field_crop_metrics.",
		strPtr("dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD"),
		&recalcBVal,
		strPtr("v4_report.dashboard_management_balance"),
		strPtr("Segundo recálculo: costos directos ejecutados del Dashboard, calculado por funciones SSOT."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 3: Informe por campo → Informe Generales (Costos Directos)
// System: ∑(field_crop_metrics.total_direct_costs)
// RecalcA: summary_results.TotalDirectCostsUsd
// RecalcB: ∑(lot.cost×ha)
// =====================================================
func (u *UseCases) control3InformeCampoVsInformeGenerales(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sumFieldCropDirectCosts(sd.fieldCropMetrics)

	recalcA := decimal.Zero
	if len(sd.summaryResults) > 0 {
		recalcA = sd.summaryResults[0].TotalDirectCostsUsd
	}

	recalcBVal := sumLotsCostDirecto(sd.lots)

	return buildCheck(
		3,
		"Costos directos ejecutados",
		"Informe por campo vs Informe Generales vs Lotes",
		"∑(field_crop.total_direct_costs) = summary_results.TotalDirectCostsUsd = ∑(lot.cost×ha)",
		"∑(field_crop_metrics.total_direct_costs_usd)",
		systemValue,
		"v4_report.field_crop_metrics",
		"Total de costos directos del Informe de Resultado por campo. Cada combinación campo+cultivo tiene sus costos calculados en field_crop_aggregated.",
		"summary_results[0].TotalDirectCostsUsd",
		recalcA,
		"v4_report.summary_results",
		"Recálculo: total de costos directos del resumen general, agrega desde field_crop_metrics → summary_results.",
		strPtr("∑(lot_list.cost_usd_per_ha × lot_list.hectares)"),
		&recalcBVal,
		strPtr("v4_report.lot_list"),
		strPtr("Segundo recálculo: costo por hectárea de cada lote × hectáreas, camino lot_base_costs → lot_metrics → lot_list."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 4: Informe Generales → Dashboard (Costos Directos)
// System: summary_results.TotalDirectCostsUsd
// RecalcA: dashboard.DirectCostsExecutedUSD
// RecalcB: ∑(lot.cost×ha)
// =====================================================
func (u *UseCases) control4InformeGeneralesVsDashboard(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := decimal.Zero
	if len(sd.summaryResults) > 0 {
		systemValue = sd.summaryResults[0].TotalDirectCostsUsd
	}

	recalcA := sd.dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD
	recalcBVal := sumLotsCostDirecto(sd.lots)

	return buildCheck(
		4,
		"Costos directos ejecutados",
		"Informe Generales vs Dashboard vs Lotes",
		"summary_results.TotalDirectCostsUsd = dashboard.DirectCostsExecutedUSD = ∑(lot.cost×ha)",
		"summary_results[0].TotalDirectCostsUsd",
		systemValue,
		"v4_report.summary_results",
		"Total de costos directos del Informe de Resultado Generales. Agrega los costos de field_crop_metrics por cultivo.",
		"dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD",
		recalcA,
		"v4_report.dashboard_management_balance",
		"Recálculo: costos directos ejecutados del Dashboard, calculado por funciones SSOT.",
		strPtr("∑(lot_list.cost_usd_per_ha × lot_list.hectares)"),
		&recalcBVal,
		strPtr("v4_report.lot_list"),
		strPtr("Segundo recálculo: costo por hectárea de cada lote × hectáreas, camino lot_base_costs → lot_metrics → lot_list."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 5: Componentes → Dashboard Total Invertido
// System: dashboard.TotalsRow.InvestedUSD
// RecalcA: Semillas + Agroquímicos + Fertilizantes + Labores
// =====================================================
func (u *UseCases) control5LaboresInsumosVsDashboard(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sd.dashboardData.ManagementBalance.TotalsRow.InvestedUSD

	summary := sd.dashboardData.ManagementBalance.Summary
	recalcA := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD).
		Add(summary.LaboresInvertidosUSD)

	return buildCheck(
		5,
		"Invertidos total",
		"Dashboard total invertido vs suma de componentes",
		"TotalsRow.InvestedUSD = Semillas + Agroquímicos + Fertilizantes + Labores",
		"dashboard.ManagementBalance.TotalsRow.InvestedUSD",
		systemValue,
		"Dashboard TotalsRow",
		"Total invertido que muestra la fila de totales del Cuadro de Gestión.",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD",
		recalcA,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Recálculo: suma los 4 componentes invertidos del Cuadro de Gestión.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 6: Dashboard Labores → Informe de Aportes
// System: dashboard.LaboresInvertidosUSD
// RecalcA: ∑(contributions: Labores Generales + Siembra + Riego)
// =====================================================
func (u *UseCases) control6LaboresVsAportes(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sd.dashboardData.ManagementBalance.Summary.LaboresInvertidosUSD

	recalcA := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Labores Generales" || cat.Label == "Siembra" || cat.Label == "Riego" {
			recalcA = recalcA.Add(cat.TotalUsd)
		}
	}

	return buildCheck(
		6,
		"Inversión en labores",
		"Dashboard labores vs Informe de Aportes",
		"Summary.LaboresInvertidosUSD = ∑(contributions: Labores Generales, Siembra, Riego)",
		"dashboard.ManagementBalance.Summary.LaboresInvertidosUSD",
		systemValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Labores invertidas del Cuadro de Gestión. Suma labores de órdenes de trabajo con estado invertido.",
		"∑(contribution_categories.total_usd) labels: Labores Generales, Siembra, Riego",
		recalcA,
		"v4_report.investor_contribution_data.contribution_categories",
		"Recálculo: suma categorías de aportes de inversores para labores, camino independiente al dashboard.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 7: Dashboard Insumos → Informe de Aportes
// System: dashboard.(Semillas + Agroquímicos + Fertilizantes)
// RecalcA: ∑(contributions: Semilla + Agroquímicos + Fertilizantes)
// =====================================================
func (u *UseCases) control7InsumosVsAportes(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	summary := sd.dashboardData.ManagementBalance.Summary
	systemValue := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD)

	recalcA := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Semilla" || cat.Label == "Agroquímicos" || cat.Label == "Fertilizantes" {
			recalcA = recalcA.Add(cat.TotalUsd)
		}
	}

	return buildCheck(
		7,
		"Inversión en insumos",
		"Dashboard insumos vs Informe de Aportes",
		"(Semillas+Agroquímicos+Fertilizantes)InvertidosUSD = ∑(contributions: Semilla, Agroquímicos, Fertilizantes)",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD",
		systemValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Semillas, agroquímicos y fertilizantes invertidos del Cuadro de Gestión.",
		"∑(contribution_categories.total_usd) labels: Semilla, Agroquímicos, Fertilizantes",
		recalcA,
		"v4_report.investor_contribution_data.contribution_categories",
		"Recálculo: suma categorías de aportes de inversores para insumos.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 8: Lotes Admin → Informe de Aportes
// System: contribution "Administración y Estructura"
// RecalcA: ∑(lot.admin_cost × hectares)
// =====================================================
func (u *UseCases) control8LotesAdminVsAportes(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Administración y Estructura" {
			systemValue = cat.TotalUsd
			break
		}
	}

	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.AdminCost.Mul(lot.Hectares))
	}

	return buildCheck(
		8,
		"Administración y Estructura",
		"Informe de Aportes vs Lotes admin",
		"contribution(Administración y Estructura).TotalUsd = ∑(lot.admin_cost × ha)",
		"contribution_categories(Administración y Estructura).total_usd",
		systemValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Categoría Administración y Estructura del Informe de Aportes.",
		"∑(lot_list.admin_cost_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo: admin prorrateado por lote desde lot_metrics.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 9: Lotes Arriendo → Informe de Aportes
// System: contribution "Arriendo Capitalizable"
// RecalcA: ∑(lot.rent_per_ha × hectares)
// =====================================================
func (u *UseCases) control9LotesArriendoVsAportes(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Arriendo Capitalizable" {
			systemValue = cat.TotalUsd
			break
		}
	}

	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.RentPerHa.Mul(lot.Hectares))
	}

	return buildCheck(
		9,
		"Arriendo Capitalizable",
		"Informe de Aportes vs Lotes arriendo",
		"contribution(Arriendo Capitalizable).TotalUsd = ∑(lot.rent_per_ha × ha)",
		"contribution_categories(Arriendo Capitalizable).total_usd",
		systemValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Categoría Arriendo Capitalizable del Informe de Aportes.",
		"∑(lot_list.rent_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo: arriendo por lote desde lot_metrics.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 10: Lotes Ingreso Neto → Resumen
// System: ∑(summary_results.NetIncomeUsd)
// RecalcA: ∑(lot.income_net_per_ha × hectares)
// =====================================================
func (u *UseCases) control10LotesIngresoNetoVsResumen(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := decimal.Zero
	for _, result := range sd.summaryResults {
		systemValue = systemValue.Add(result.NetIncomeUsd)
	}

	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.IncomeNetPerHa.Mul(lot.Hectares))
	}

	return buildCheck(
		10,
		"Ingreso Neto",
		"Resumen de resultados vs Lotes ingreso neto",
		"∑(summary_results.NetIncomeUsd) = ∑(lot.income_net_per_ha × ha)",
		"∑(summary_results.net_income_usd)",
		systemValue,
		"v4_report.summary_results",
		"Ingreso neto total del Informe de Resultado Generales.",
		"∑(lot_list.income_net_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo: ingreso neto por hectárea de cada lote × hectáreas.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 11: Lotes → Informe por cultivo (Resultado Operativo)
// System: ∑(lot.operating_result × ha)
// RecalcA: ∑(field_crop.operating_result × surface)
// RecalcB: dashboard.OperatingResult.ResultUSD
// =====================================================
func (u *UseCases) control11LotesResultadoVsInformeCultivo(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sumLotsOperatingResult(sd.lots)
	recalcA := sumFieldCropOperatingResult(sd.fieldCropMetrics)
	recalcBVal := sd.dashboardData.Metrics.OperatingResult.ResultUSD

	return buildCheck(
		11,
		"Resultado operativo total",
		"Lotes vs Informe por cultivo vs Dashboard",
		"∑(lot.operating_result×ha) = ∑(field_crop.operating_result×surface) = dashboard.OperatingResult",
		"∑(lot_list.operating_result_per_ha × lot_list.hectares)",
		systemValue,
		"v4_report.lot_list",
		"Resultado operativo total desde tabla de lotes. Cada lote: ingreso neto - activo total.",
		"∑(field_crop_metrics.operating_result_usd_ha × field_crop_metrics.surface_ha)",
		recalcA,
		"v4_report.field_crop_metrics",
		"Recálculo: resultado operativo por campo/cultivo desde field_crop_aggregated.",
		strPtr("dashboard.Metrics.OperatingResult.ResultUSD"),
		&recalcBVal,
		strPtr("v4_report.dashboard_management_balance"),
		strPtr("Segundo recálculo: resultado operativo de la tarjeta del Dashboard."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 12: Lotes → Informe Generales (Resultado Operativo)
// System: ∑(lot.operating_result × ha)
// RecalcA: summary_results.TotalOperatingResultUsd
// RecalcB: dashboard.OperatingResult.ResultUSD
// =====================================================
func (u *UseCases) control12LotesResultadoVsInformeGenerales(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sumLotsOperatingResult(sd.lots)

	recalcA := decimal.Zero
	if len(sd.summaryResults) > 0 {
		recalcA = sd.summaryResults[0].TotalOperatingResultUsd
	}

	recalcBVal := sd.dashboardData.Metrics.OperatingResult.ResultUSD

	return buildCheck(
		12,
		"Resultado operativo total",
		"Lotes vs Informe Generales vs Dashboard",
		"∑(lot.operating_result×ha) = summary_results.TotalOperatingResultUsd = dashboard.OperatingResult",
		"∑(lot_list.operating_result_per_ha × lot_list.hectares)",
		systemValue,
		"v4_report.lot_list",
		"Resultado operativo total desde tabla de lotes.",
		"summary_results[0].TotalOperatingResultUsd",
		recalcA,
		"v4_report.summary_results",
		"Recálculo: resultado operativo del resumen general, agrega por cultivo.",
		strPtr("dashboard.Metrics.OperatingResult.ResultUSD"),
		&recalcBVal,
		strPtr("v4_report.dashboard_management_balance"),
		strPtr("Segundo recálculo: resultado operativo de la tarjeta del Dashboard."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 13: Lotes → Dashboard (Resultado Operativo)
// System: dashboard.OperatingResult.ResultUSD
// RecalcA: ∑(lot.operating_result × ha)
// RecalcB: summary_results.TotalOperatingResultUsd
// =====================================================
func (u *UseCases) control13LotesResultadoVsDashboard(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	systemValue := sd.dashboardData.Metrics.OperatingResult.ResultUSD
	recalcA := sumLotsOperatingResult(sd.lots)

	var recalcBCalc, recalcBSrc, recalcBMeaning *string
	var recalcBVal *decimal.Decimal
	if len(sd.summaryResults) > 0 {
		v := sd.summaryResults[0].TotalOperatingResultUsd
		recalcBCalc = strPtr("summary_results[0].TotalOperatingResultUsd")
		recalcBVal = &v
		recalcBSrc = strPtr("v4_report.summary_results")
		recalcBMeaning = strPtr("Segundo recálculo: resultado operativo del resumen general.")
	}

	return buildCheck(
		13,
		"Resultado operativo total",
		"Dashboard vs Lotes vs Informe Generales",
		"dashboard.OperatingResult.ResultUSD = ∑(lot.operating_result×ha) = summary_results.TotalOperatingResultUsd",
		"dashboard.Metrics.OperatingResult.ResultUSD",
		systemValue,
		"v4_report.dashboard_management_balance",
		"Resultado operativo de la tarjeta del Dashboard. Ingreso neto menos costos totales activos.",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo: resultado operativo por lote × hectáreas.",
		recalcBCalc, recalcBVal, recalcBSrc, recalcBMeaning,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 14: Stock → Dashboard
// System: dashboard.StockUSD
// RecalcA: (Invertido - Ejecutado)
// =====================================================
func (u *UseCases) control14StockVsDashboard(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	summary := sd.dashboardData.ManagementBalance.Summary
	systemValue := summary.StockUSD

	invertido := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD).
		Add(summary.LaboresInvertidosUSD)
	ejecutado := summary.DirectCostsExecutedUSD
	recalcA := invertido.Sub(ejecutado)

	return buildCheck(
		14,
		"Stock",
		"Dashboard stock vs cálculo Invertido - Ejecutado",
		"Summary.StockUSD = (Semillas+Agroquímicos+Fertilizantes+Labores) - DirectCostsExecuted",
		"dashboard.ManagementBalance.Summary.StockUSD",
		systemValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Stock del Cuadro de Gestión: diferencia entre invertido y ejecutado.",
		"(SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD) - DirectCostsExecutedUSD",
		recalcA,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Recálculo: toma los 4 componentes invertidos y resta los costos directos ejecutados.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// HELPERS
// =====================================================

func strPtr(s string) *string { return &s }

// buildCheck construye un IntegrityCheck con SystemValue / RecalcA / RecalcB (opcional)
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
	recalcBCalculation *string,
	recalcBValue *decimal.Decimal,
	recalcBSource *string,
	recalcBMeaning *string,
	tolerance decimal.Decimal,
) domain.IntegrityCheck {
	differenceA := systemValue.Sub(recalcAValue)
	status := "OK"

	if differenceA.Abs().GreaterThan(tolerance) {
		status = "ERROR"
	}

	var differenceB *decimal.Decimal
	if recalcBValue != nil {
		diff := systemValue.Sub(*recalcBValue)
		differenceB = &diff
		if diff.Abs().GreaterThan(tolerance) {
			status = "ERROR"
		}
	}

	return domain.IntegrityCheck{
		ControlNumber:      controlNumber,
		DataToVerify:       dataToVerify,
		Description:        description,
		ControlRule:        controlRule,
		SystemCalculation:  systemCalculation,
		SystemValue:        systemValue,
		SystemSource:       systemSource,
		SystemMeaning:      systemMeaning,
		RecalcACalculation: recalcACalculation,
		RecalcAValue:       recalcAValue,
		RecalcASource:      recalcASource,
		RecalcAMeaning:     recalcAMeaning,
		RecalcBCalculation: recalcBCalculation,
		RecalcBValue:       recalcBValue,
		RecalcBSource:      recalcBSource,
		RecalcBMeaning:     recalcBMeaning,
		DifferenceA:        differenceA,
		DifferenceB:        differenceB,
		Status:             status,
		Tolerance:          tolerance,
	}
}
