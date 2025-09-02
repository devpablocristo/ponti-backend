package models

import (
	"time"

	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

// DashboardModel representa el modelo de base de datos para dashboard
type DashboardModel struct {
	ID int64

	sharedmodels.Base
}

// DashboardDataModel representa el modelo de datos del dashboard consolidado
type DashboardDataModel struct {
	// Métricas de siembra
	SowingHectares        decimal.Decimal `db:"sowing_hectares"`
	SowingTotalHectares   decimal.Decimal `db:"sowing_total_hectares"`
	SowingProgressPercent decimal.Decimal `db:"sowing_progress_percent"`

	// Métricas de cosecha
	HarvestHectares        decimal.Decimal `db:"harvest_hectares"`
	HarvestTotalHectares   decimal.Decimal `db:"harvest_total_hectares"`
	HarvestProgressPercent decimal.Decimal `db:"harvest_progress_percent"`

	// Métricas de costos
	CostsExecutedUSD    decimal.Decimal `db:"costs_executed_usd"`
	CostsBudgetUSD      decimal.Decimal `db:"costs_budget_usd"`
	CostsProgressPct    decimal.Decimal `db:"costs_progress_pct"`
	ExecutedLaborsUSD   decimal.Decimal `db:"executed_labors_usd"`
	ExecutedSuppliesUSD decimal.Decimal `db:"executed_supplies_usd"`
	BudgetCostUSD       decimal.Decimal `db:"budget_cost_usd"`

	// Resultado operativo
	OperatingIncomeUSD     decimal.Decimal `db:"operating_income_usd"`
	OperatingTotalCostsUSD decimal.Decimal `db:"operating_total_costs_usd"`
	OperatingResultUSD     decimal.Decimal `db:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `db:"operating_result_pct"`

	// Balance de gestión - Semilla
	SemillaEjecutadosUSD decimal.Decimal `db:"semilla_ejecutados_usd"`
	SemillaInvertidosUSD decimal.Decimal `db:"semilla_invertidos_usd"`
	SemillaStockUSD      decimal.Decimal `db:"semilla_stock_usd"`

	// Balance de gestión - Insumos
	InsumosEjecutadosUSD decimal.Decimal `db:"insumos_ejecutados_usd"`
	InsumosInvertidosUSD decimal.Decimal `db:"insumos_invertidos_usd"`
	InsumosStockUSD      decimal.Decimal `db:"insumos_stock_usd"`

	// Balance de gestión - Labores
	LaboresEjecutadosUSD decimal.Decimal `db:"labores_ejecutados_usd"`
	LaboresInvertidosUSD decimal.Decimal `db:"labores_invertidos_usd"`
	LaboresStockUSD      decimal.Decimal `db:"labores_stock_usd"`

	// Balance de gestión - Arriendo
	ArriendoInvertidosUSD decimal.Decimal `db:"arriendo_invertidos_usd"`

	// Balance de gestión - Estructura
	EstructuraInvertidosUSD decimal.Decimal `db:"estructura_invertidos_usd"`

	// Totales del Balance de Gestión
	CostosDirectosEjecutados decimal.Decimal `db:"costos_directos_ejecutados_usd"`
	CostosDirectosInvertidos decimal.Decimal `db:"costos_directos_invertidos_usd"`
	CostosDirectosStock      decimal.Decimal `db:"costos_directos_stock_usd"`
}

// SowingProgressModel representa el modelo de avance de siembra
type SowingProgressModel struct {
	Hectares      *decimal.Decimal `db:"sowing_hectares"`
	TotalHectares *decimal.Decimal `db:"sowing_total_hectares"`
	ProgressPct   *decimal.Decimal `db:"sowing_progress_pct"`
}

// HarvestProgressModel representa el modelo de avance de cosecha
type HarvestProgressModel struct {
	Hectares      *decimal.Decimal `db:"harvest_hectares"`
	TotalHectares *decimal.Decimal `db:"harvest_total_hectares"`
	ProgressPct   *decimal.Decimal `db:"harvest_progress_pct"`
}

// CostsProgressModel representa el modelo de avance de costos
type CostsProgressModel struct {
	ExecutedLaborsUSD   *decimal.Decimal `db:"executed_labors_usd"`
	ExecutedSuppliesUSD *decimal.Decimal `db:"executed_supplies_usd"`
	ExecutedCostsUSD    *decimal.Decimal `db:"executed_costs_usd"`
	BudgetCostUSD       *decimal.Decimal `db:"budget_cost_usd"`
	BudgetTotalUSD      *decimal.Decimal `db:"budget_total_usd"`
	ProgressPct         *decimal.Decimal `db:"costs_progress_pct"`
}

// ContributionsProgressModel representa el modelo de avance de aportes
type ContributionsProgressModel struct {
	InvestorID               *int64           `db:"investor_id"`
	InvestorName             *string          `db:"investor_name"`
	InvestorPercentage       *decimal.Decimal `db:"investor_percentage_pct"`
	ContributionsProgressPct *decimal.Decimal `db:"contributions_progress_pct"`
}

// OperatingResultModel representa el modelo de resultado operativo
type OperatingResultModel struct {
	IncomeUSD     *decimal.Decimal `db:"income_usd"`
	TotalCostsUSD *decimal.Decimal `db:"operating_result_total_costs_usd"`
	ResultUSD     *decimal.Decimal `db:"operating_result_usd"`
	ResultPct     *decimal.Decimal `db:"operating_result_pct"`
}

// ManagementBalanceModel representa el modelo de balance de gestión
type ManagementBalanceModel struct {
	Summary   *ManagementBalanceSummary
	Breakdown []ManagementBalanceBreakdown
	TotalsRow *ManagementBalanceTotals
}

// ManagementBalanceSummary representa el resumen del balance de gestión
type ManagementBalanceSummary struct {
	IncomeUSD              decimal.Decimal `db:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `db:"costos_directos_ejecutados_usd"`
	DirectCostsInvestedUSD decimal.Decimal `db:"costos_directos_invertidos_usd"`
	StockUSD               decimal.Decimal `db:"costos_directos_stock_usd"`
	RentUSD                decimal.Decimal `db:"arriendo_invertidos_usd"`
	StructureUSD           decimal.Decimal `db:"estructura_invertidos_usd"`
	OperatingResultUSD     decimal.Decimal `db:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `db:"operating_result_pct"`
}

// ManagementBalanceBreakdown representa el desglose del balance por categoría
type ManagementBalanceBreakdown struct {
	Category    string          `db:"category"`
	ExecutedUSD decimal.Decimal `db:"executed_usd"`
	InvestedUSD decimal.Decimal `db:"invested_usd"`
	StockUSD    decimal.Decimal `db:"stock_usd"`
}

// ManagementBalanceTotals representa los totales del balance de gestión
type ManagementBalanceTotals struct {
	TotalExecutedUSD decimal.Decimal `db:"total_executed_usd"`
	TotalInvestedUSD decimal.Decimal `db:"total_invested_usd"`
	TotalStockUSD    decimal.Decimal `db:"total_stock_usd"`
}

// CropIncidenceModel representa el modelo de incidencia de cultivos
type CropIncidenceModel struct {
	Name         string          `db:"crop_name"`
	Hectares     decimal.Decimal `db:"crop_hectares"`
	IncidencePct decimal.Decimal `db:"incidence_pct"`
	CostPerHa    decimal.Decimal `db:"cost_per_ha_usd"`
}

// InvestorContributionModel representa el modelo de contribución de inversores
type InvestorContributionModel struct {
	InvestorID   int64           `db:"investor_id"`
	InvestorName string          `db:"investor_name"`
	Percentage   decimal.Decimal `db:"investor_percentage_pct"`
}

// OperationalIndicatorModel representa el modelo de indicadores operativos
type OperationalIndicatorModel struct {
	PrimeraOrdenFecha  *time.Time `db:"primera_orden_fecha"`
	PrimeraOrdenID     *int64     `db:"primera_orden_id"`
	UltimaOrdenFecha   *time.Time `db:"ultima_orden_fecha"`
	UltimaOrdenID      *int64     `db:"ultima_orden_id"`
	ArqueoStockFecha   *time.Time `db:"arqueo_stock_fecha"`
	CierreCampanaFecha *time.Time `db:"cierre_campana_fecha"`
}
