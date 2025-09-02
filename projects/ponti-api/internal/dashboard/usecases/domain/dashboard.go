package domain

import (
	"time"

	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

// DashboardFilter representa los filtros para obtener datos del dashboard
type DashboardFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}

// Dashboard representa la entidad principal del dashboard
type Dashboard struct {
	ID int64

	shareddomain.Base
}

// DashboardData representa los datos consolidados del dashboard
type DashboardData struct {
	Metrics               *DashboardMetrics
	ManagementBalance     *DashboardManagementBalance
	CropIncidence         *DashboardCropIncidence
	OperationalIndicators *DashboardOperationalIndicators
}

// DashboardMetrics representa las métricas del dashboard
type DashboardMetrics struct {
	Sowing                *DashboardSowing
	Harvest               *DashboardHarvest
	Costs                 *DashboardCosts
	InvestorContributions *DashboardInvestorContributions
	OperatingResult       *DashboardOperatingResult
}

// DashboardSowing representa la métrica de siembra
type DashboardSowing struct {
	ProgressPct   decimal.Decimal
	Hectares      decimal.Decimal
	TotalHectares decimal.Decimal
}

// DashboardHarvest representa la métrica de cosecha
type DashboardHarvest struct {
	ProgressPct   decimal.Decimal
	Hectares      decimal.Decimal
	TotalHectares decimal.Decimal
}

// DashboardCosts representa la métrica de costos
type DashboardCosts struct {
	ProgressPct decimal.Decimal
	ExecutedUSD decimal.Decimal
	BudgetUSD   decimal.Decimal
}

// DashboardInvestorContributions representa la métrica de contribuciones
type DashboardInvestorContributions struct {
	ProgressPct decimal.Decimal
	Breakdown   []DashboardInvestorBreakdown
}

// DashboardInvestorBreakdown representa el desglose de contribuciones por inversor
type DashboardInvestorBreakdown struct {
	InvestorID   int64
	InvestorName string
	PercentPct   decimal.Decimal
}

// DashboardOperatingResult representa la métrica de resultado operativo
type DashboardOperatingResult struct {
	ProgressPct   decimal.Decimal
	ResultUSD     decimal.Decimal
	TotalCostsUSD decimal.Decimal
}

// DashboardManagementBalance representa el balance de gestión
type DashboardManagementBalance struct {
	Summary   *DashboardBalanceSummary
	Breakdown []DashboardBalanceBreakdown
	TotalsRow *DashboardBalanceTotals
}

// DashboardBalanceSummary representa el resumen del balance
type DashboardBalanceSummary struct {
	IncomeUSD              decimal.Decimal
	DirectCostsExecutedUSD decimal.Decimal
	DirectCostsInvestedUSD decimal.Decimal
	StockUSD               decimal.Decimal
	RentUSD                decimal.Decimal
	StructureUSD           decimal.Decimal
	OperatingResultUSD     decimal.Decimal
	OperatingResultPct     decimal.Decimal
	SemillaCostUSD         decimal.Decimal
	InsumosCostUSD         decimal.Decimal
	LaboresCostUSD         decimal.Decimal
}

// DashboardBalanceBreakdown representa el desglose del balance por categoría
type DashboardBalanceBreakdown struct {
	Label       string
	ExecutedUSD decimal.Decimal
	InvestedUSD decimal.Decimal
	StockUSD    *decimal.Decimal
}

// DashboardBalanceTotals representa los totales del balance
type DashboardBalanceTotals struct {
	ExecutedUSD decimal.Decimal
	InvestedUSD decimal.Decimal
	StockUSD    decimal.Decimal
}

// DashboardCropIncidence representa la incidencia de costos por cultivo
type DashboardCropIncidence struct {
	Crops []DashboardCropBreakdown
	Total *DashboardCropTotal
}

// DashboardCropBreakdown representa el desglose de costos por cultivo
type DashboardCropBreakdown struct {
	Name         string
	Hectares     decimal.Decimal
	RotationPct  decimal.Decimal
	CostUSDPerHa decimal.Decimal
	IncidencePct decimal.Decimal
}

// DashboardCropTotal representa los totales de cultivos
type DashboardCropTotal struct {
	Hectares          decimal.Decimal
	RotationPct       decimal.Decimal
	CostUSDPerHectare decimal.Decimal
}

// DashboardOperationalIndicators representa los indicadores operativos
type DashboardOperationalIndicators struct {
	Cards []DashboardOperationalCard
}

// DashboardOperationalCard representa una tarjeta de indicador operativo
type DashboardOperationalCard struct {
	Key         string
	Title       string
	Date        *time.Time
	WorkorderID *int64
}
