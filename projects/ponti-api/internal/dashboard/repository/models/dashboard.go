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

// DashboardDataModel representa el modelo de datos del dashboard desde la vista
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
