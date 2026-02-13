// Package dataintegrity implementa casos de uso para validar la coherencia de datos
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
//
// ESTOS CÁLCULOS SON CRÍTICOS Y NO DEBEN ALTERARSE A MENOS QUE SE RECIBA
// UNA ORDEN DIRECTA Y CLARA DEL USUARIO.
//
// REGLAS INVIOLABLES:
// - NUNCA modificar los cálculos System/RecalcA/RecalcB sin autorización explícita
// - NUNCA cambiar las tolerancias sin autorización explícita
// - NUNCA alterar la lógica de los 9 controles sin autorización explícita
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

// sharedData cachea datos compartidos entre controles para reducir round-trips a DB.
type sharedData struct {
	lots             []lotDomain.LotTable
	dashboardData    *dashboardDomain.DashboardData
	fieldCropMetrics []reportDomain.FieldCropMetric
	summaryResults   []reportDomain.SummaryResults
	investorReport   *reportDomain.InvestorContributionReport
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

	return sd, nil
}

// CheckCostsCoherence valida la coherencia de costos con 9 controles individuales.
// Cada control compara: SystemValue (lo que el sistema muestra) vs RecalcA y RecalcB (recálculos independientes).
//
// ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️
// ESTA FUNCIÓN CONTIENE LOS 9 CONTROLES CRÍTICOS DE INTEGRIDAD DE DATOS.
// NUNCA ALTERAR SIN AUTORIZACIÓN EXPLÍCITA DEL USUARIO.
func (u *UseCases) CheckCostsCoherence(ctx context.Context, filter domain.CostsCheckFilter) (*domain.IntegrityReport, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Pre-fetch datos compartidos una sola vez (reduce ~25+ queries a ~5)
	sd, err := u.fetchSharedData(ctx, filter.ProjectID)
	if err != nil {
		return nil, err
	}

	checks := make([]domain.IntegrityCheck, 9)
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

	// CONTROL 1: Costos directos ejecutados (consolida antiguos 1,2,3,4)
	run(0, 1, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control1CostosDirectos(c, sd)
	})
	// CONTROL 2: Invertidos total (antiguo 5)
	run(1, 2, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control2InvertidosTotal(c, sd)
	})
	// CONTROL 3: Labores invertidos (antiguo 6)
	run(2, 3, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control3LaboresInvertidos(c, sd)
	})
	// CONTROL 4: Insumos invertidos (antiguo 7)
	run(3, 4, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control4InsumosInvertidos(c, sd)
	})
	// CONTROL 5: Administración (antiguo 8)
	run(4, 5, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control5Administracion(c, sd)
	})
	// CONTROL 6: Arriendo (antiguo 9)
	run(5, 6, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control6Arriendo(c, sd)
	})
	// CONTROL 7: Ingreso Neto (antiguo 10)
	run(6, 7, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control7IngresoNeto(c, sd)
	})
	// CONTROL 8: Resultado operativo (consolida antiguos 11,12,13)
	run(7, 8, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control8ResultadoOperativo(c, sd)
	})
	// CONTROL 9: Stock (antiguo 14)
	run(8, 9, func(c context.Context) (domain.IntegrityCheck, error) {
		return u.control9Stock(c, sd)
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
// CONTROL 1: Costos Directos Ejecutados
// System: dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD
// RecalcA: ∑(lot_list.cost_usd_per_ha × lot_list.hectares)
// RecalcB: summary_results[0].TotalDirectCostsUsd
// =====================================================
func (u *UseCases) control1CostosDirectos(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: lo que el dashboard muestra como costos directos ejecutados
	systemValue := sd.dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD

	// RECALC A: sumar desde lotes (cost_per_ha × hectares)
	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.CostUsdPerHa.Mul(lot.Hectares))
	}

	// RECALC B: total desde summary_results
	var recalcBCalc *string
	var recalcBVal *decimal.Decimal
	var recalcBSrc *string
	if len(sd.summaryResults) > 0 {
		v := sd.summaryResults[0].TotalDirectCostsUsd
		recalcBCalc = strPtr("summary_results[0].TotalDirectCostsUsd")
		recalcBVal = &v
		recalcBSrc = strPtr("v4_report.summary_results")
	}

	return buildCheck(
		1,
		"Costos directos ejecutados",
		"Compara costos directos ejecutados entre dashboard, lotes y resumen.",
		"dashboard.DirectCostsExecutedUSD = ∑(lot.cost×ha) = summary_results.TotalDirectCostsUsd",
		"dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD",
		systemValue,
		"v4_report.dashboard_management_balance",
		"Valor de costos directos ejecutados que muestra el Cuadro de Gestión del Dashboard. Se calcula en la vista dashboard_management_balance sumando los costos de labores e insumos ejecutados por lote, usando las funciones SSOT.",
		"∑(lot_list.cost_usd_per_ha × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo independiente: toma el costo por hectárea de cada lote (de la vista lot_list) y lo multiplica por sus hectáreas, luego suma todos los lotes. Usa el camino workorder_metrics_raw → lot_base_costs → lot_metrics → lot_list.",
		recalcBCalc, recalcBVal, recalcBSrc,
		strPtr("Segundo recálculo independiente: toma el total de costos directos del resumen general (vista summary_results), que agrega por cultivo usando field_crop_aggregated → field_crop_metrics → summary_results."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 2: Invertidos Total
// System: dashboard.ManagementBalance.TotalsRow.InvestedUSD
// RecalcA: (Sem + Agro + Fert + Lab) InvertidosUSD del summary
// =====================================================
func (u *UseCases) control2InvertidosTotal(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: total invertido que muestra el dashboard
	systemValue := sd.dashboardData.ManagementBalance.TotalsRow.InvestedUSD

	// RECALC A: suma de componentes desde dashboard summary
	summary := sd.dashboardData.ManagementBalance.Summary
	recalcA := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD).
		Add(summary.LaboresInvertidosUSD)

	return buildCheck(
		2,
		"Invertidos total",
		"Valida que el total invertido sea la suma de sus componentes.",
		"TotalsRow.InvestedUSD = Semillas + Agroquímicos + Fertilizantes + Labores",
		"dashboard.ManagementBalance.TotalsRow.InvestedUSD",
		systemValue,
		"Dashboard TotalsRow",
		"Total invertido que muestra la fila de totales del Cuadro de Gestión. Es la suma precalculada en la vista dashboard_management_balance como total de todos los rubros invertidos.",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD",
		recalcA,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Recálculo independiente: suma los 4 componentes individuales del mismo Cuadro de Gestión (semillas + agroquímicos + fertilizantes + labores invertidos). Si difiere del total, hay un error en la agregación de la vista.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 3: Labores Invertidos
// System: dashboard.ManagementBalance.Summary.LaboresInvertidosUSD
// RecalcA: ∑(investor_contributions: Labores Generales + Siembra + Riego)
// =====================================================
func (u *UseCases) control3LaboresInvertidos(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: labores invertidas que muestra el dashboard
	systemValue := sd.dashboardData.ManagementBalance.Summary.LaboresInvertidosUSD

	// RECALC A: desde informe de aportes
	recalcA := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Labores Generales" || cat.Label == "Siembra" || cat.Label == "Riego" {
			recalcA = recalcA.Add(cat.TotalUsd)
		}
	}

	return buildCheck(
		3,
		"Labores invertidos",
		"Valida labores invertidas entre dashboard y aportes.",
		"Summary.LaboresInvertidosUSD = ∑(contributions: Labores Generales, Siembra, Riego)",
		"dashboard.ManagementBalance.Summary.LaboresInvertidosUSD",
		systemValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Monto de labores invertidas que muestra el Cuadro de Gestión del Dashboard. Se calcula en la vista dashboard_management_balance sumando el costo de labores de todas las órdenes de trabajo con estado 'invertido'.",
		"∑(contribution_categories.total_usd) labels: Labores Generales, Siembra, Riego",
		recalcA,
		"v4_report.investor_contribution_data.contribution_categories",
		"Recálculo independiente: suma las categorías de aportes de inversores que corresponden a labores (Labores Generales + Siembra + Riego) desde la vista investor_contribution_data. Camino completamente distinto al dashboard.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 4: Insumos Invertidos
// System: dashboard.(Sem+Agro+Fert)InvertidosUSD
// RecalcA: ∑(investor_contributions: Semilla + Agroquímicos + Fertilizantes)
// =====================================================
func (u *UseCases) control4InsumosInvertidos(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: insumos invertidos que muestra el dashboard
	summary := sd.dashboardData.ManagementBalance.Summary
	systemValue := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD)

	// RECALC A: desde informe de aportes
	recalcA := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Semilla" || cat.Label == "Agroquímicos" || cat.Label == "Fertilizantes" {
			recalcA = recalcA.Add(cat.TotalUsd)
		}
	}

	return buildCheck(
		4,
		"Insumos invertidos",
		"Valida insumos invertidos entre dashboard y aportes.",
		"(Semillas+Agroquímicos+Fertilizantes)InvertidosUSD = ∑(contributions: Semilla, Agroquímicos, Fertilizantes)",
		"SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD",
		systemValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Suma de los 3 rubros de insumos invertidos que muestra el Cuadro de Gestión: semillas, agroquímicos y fertilizantes. Cada rubro se calcula en dashboard_management_balance desde las órdenes de trabajo con estado 'invertido'.",
		"∑(contribution_categories.total_usd) labels: Semilla, Agroquímicos, Fertilizantes",
		recalcA,
		"v4_report.investor_contribution_data.contribution_categories",
		"Recálculo independiente: suma las categorías de aportes de inversores que corresponden a insumos (Semilla + Agroquímicos + Fertilizantes) desde la vista investor_contribution_data.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 5: Administración y Estructura
// System: investor_contribution "Administración y Estructura".TotalUsd
// RecalcA: ∑(lot_list.admin_cost_per_ha × hectares)
// =====================================================
func (u *UseCases) control5Administracion(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: lo que muestra el informe de aportes para Administración
	systemValue := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Administración y Estructura" {
			systemValue = cat.TotalUsd
			break
		}
	}

	// RECALC A: sumar desde lotes (admin_cost × hectares)
	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.AdminCost.Mul(lot.Hectares))
	}

	return buildCheck(
		5,
		"Administración y Estructura",
		"Valida administración entre aportes y lotes.",
		"contribution(Administración y Estructura).TotalUsd = ∑(lot.admin_cost × ha)",
		"contribution_categories(Administración y Estructura).total_usd",
		systemValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Total de la categoría 'Administración y Estructura' del Informe de Aportes. Se calcula en la vista investor_contribution_data agregando los aportes registrados para ese concepto.",
		"∑(lot_list.admin_cost_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo independiente: toma el costo de administración por hectárea de cada lote (prorrateado en lot_metrics desde el proyecto) y lo multiplica por las hectáreas del lote, luego suma todos los lotes.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 6: Arriendo Capitalizable
// System: investor_contribution "Arriendo Capitalizable".TotalUsd
// RecalcA: ∑(lot_list.rent_per_ha × hectares)
// =====================================================
func (u *UseCases) control6Arriendo(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: lo que muestra el informe de aportes para Arriendo
	systemValue := decimal.Zero
	for _, cat := range sd.investorReport.Contributions {
		if cat.Label == "Arriendo Capitalizable" {
			systemValue = cat.TotalUsd
			break
		}
	}

	// RECALC A: sumar desde lotes (rent × hectares)
	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.RentPerHa.Mul(lot.Hectares))
	}

	return buildCheck(
		6,
		"Arriendo Capitalizable",
		"Valida arriendo capitalizable entre aportes y lotes.",
		"contribution(Arriendo Capitalizable).TotalUsd = ∑(lot.rent_per_ha × ha)",
		"contribution_categories(Arriendo Capitalizable).total_usd",
		systemValue,
		"v4_report.investor_contribution_data.contribution_categories",
		"Total de la categoría 'Arriendo Capitalizable' del Informe de Aportes. Se calcula en la vista investor_contribution_data agregando los aportes registrados para arriendo.",
		"∑(lot_list.rent_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo independiente: toma el arriendo por hectárea de cada lote (asignado en lot_metrics desde el campo/proyecto) y lo multiplica por las hectáreas del lote, luego suma todos los lotes.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 7: Ingreso Neto
// System: ∑(summary_results.NetIncomeUsd)
// RecalcA: ∑(lot_list.income_net_per_ha × hectares)
// =====================================================
func (u *UseCases) control7IngresoNeto(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: ingreso neto total del resumen
	systemValue := decimal.Zero
	for _, result := range sd.summaryResults {
		systemValue = systemValue.Add(result.NetIncomeUsd)
	}

	// RECALC A: sumar desde lotes (income_net × hectares)
	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.IncomeNetPerHa.Mul(lot.Hectares))
	}

	return buildCheck(
		7,
		"Ingreso Neto",
		"Valida ingreso neto entre resumen y lotes.",
		"∑(summary_results.NetIncomeUsd) = ∑(lot.income_net_per_ha × ha)",
		"∑(summary_results.net_income_usd)",
		systemValue,
		"v4_report.summary_results",
		"Ingreso neto total del proyecto que muestra el Informe de Resultado Generales (summary_results). Se calcula sumando el ingreso neto de cada cultivo, donde cada uno usa la función SSOT income_net_total_for_lot.",
		"∑(lot_list.income_net_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo independiente: toma el ingreso neto por hectárea de cada lote (calculado en lot_metrics como ventas menos costos) y lo multiplica por las hectáreas del lote, luego suma todos los lotes.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 8: Resultado Operativo
// System: dashboard.Metrics.OperatingResult.ResultUSD
// RecalcA: ∑(lot_list.operating_result_per_ha × hectares)
// RecalcB: summary_results[0].TotalOperatingResultUsd
// =====================================================
func (u *UseCases) control8ResultadoOperativo(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	// SYSTEM: resultado operativo que muestra el dashboard
	systemValue := sd.dashboardData.Metrics.OperatingResult.ResultUSD

	// RECALC A: sumar desde lotes (operating_result × hectares)
	recalcA := decimal.Zero
	for _, lot := range sd.lots {
		recalcA = recalcA.Add(lot.OperatingResultPerHa.Mul(lot.Hectares))
	}

	// RECALC B: total desde summary_results
	var recalcBCalc *string
	var recalcBVal *decimal.Decimal
	var recalcBSrc *string
	if len(sd.summaryResults) > 0 {
		v := sd.summaryResults[0].TotalOperatingResultUsd
		recalcBCalc = strPtr("summary_results[0].TotalOperatingResultUsd")
		recalcBVal = &v
		recalcBSrc = strPtr("v4_report.summary_results")
	}

	return buildCheck(
		8,
		"Resultado operativo",
		"Compara resultado operativo entre dashboard, lotes y resumen.",
		"dashboard.OperatingResult.ResultUSD = ∑(lot.operating_result×ha) = summary_results.TotalOperatingResultUsd",
		"dashboard.Metrics.OperatingResult.ResultUSD",
		systemValue,
		"v4_report.dashboard_management_balance",
		"Resultado operativo que muestra la tarjeta del Dashboard. Se calcula en dashboard_management_balance como ingreso neto menos costos totales activos (directos + arriendo + administración).",
		"∑(lot_list.operating_result_per_ha_usd × lot_list.hectares)",
		recalcA,
		"v4_report.lot_list",
		"Recálculo independiente: toma el resultado operativo por hectárea de cada lote (calculado en lot_metrics como ingreso neto menos activo total) y lo multiplica por las hectáreas del lote, luego suma todos los lotes.",
		recalcBCalc, recalcBVal, recalcBSrc,
		strPtr("Segundo recálculo independiente: toma el resultado operativo total del resumen general (vista summary_results), que agrega por cultivo usando field_crop_metrics → summary_results."),
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// CONTROL 9: Stock
// System: dashboard.ManagementBalance.Summary.StockUSD
// RecalcA: (Invertido - Ejecutado) desde componentes del summary
// =====================================================
func (u *UseCases) control9Stock(_ context.Context, sd *sharedData) (domain.IntegrityCheck, error) {
	summary := sd.dashboardData.ManagementBalance.Summary

	// SYSTEM: stock que muestra el dashboard
	systemValue := summary.StockUSD

	// RECALC A: calcular stock esperado (Invertido - Ejecutado)
	invertido := summary.SemillasInvertidosUSD.
		Add(summary.AgroquimicosInvertidosUSD).
		Add(summary.FertilizantesInvertidosUSD).
		Add(summary.LaboresInvertidosUSD)
	ejecutado := summary.DirectCostsExecutedUSD
	recalcA := invertido.Sub(ejecutado)

	return buildCheck(
		9,
		"Stock",
		"Valida stock en dashboard.",
		"Summary.StockUSD = (Semillas+Agroquímicos+Fertilizantes+Labores)Invertidos - DirectCostsExecuted",
		"dashboard.ManagementBalance.Summary.StockUSD",
		systemValue,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Stock que muestra el Cuadro de Gestión del Dashboard. Es el valor precalculado en la vista dashboard_management_balance como la diferencia entre lo invertido total y lo ejecutado total.",
		"(SemillasInvertidosUSD + AgroquimicosInvertidosUSD + FertilizantesInvertidosUSD + LaboresInvertidosUSD) - DirectCostsExecutedUSD",
		recalcA,
		"Dashboard Summary (ManagementBalance.Summary)",
		"Recálculo independiente: toma los 4 componentes invertidos del mismo Cuadro de Gestión (semillas + agroquímicos + fertilizantes + labores) y le resta los costos directos ejecutados. Stock = Invertido - Ejecutado.",
		nil, nil, nil, nil,
		decimal.NewFromInt(1),
	), nil
}

// =====================================================
// HELPERS
// =====================================================

func strPtr(s string) *string              { return &s }

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
