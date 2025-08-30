package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type DashboardFilter struct {
	CampaignIDs []int64
	ProjectIDs  []int64
	CustomerIDs []int64
	FieldIDs    []int64
	Limit       int // 0 means no limit
	Offset      int
}

// ===== ENTIDADES PARA EL DASHBOARD =====

// DashboardSowing representa las métricas de siembra
type DashboardSowing struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// DashboardHarvest representa las métricas de cosecha
type DashboardHarvest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// DashboardCosts representa las métricas de costos
type DashboardCosts struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	BudgetUSD   decimal.Decimal `json:"budget_usd"`
}

// DashboardInvestorBreakdown representa un inversor en el breakdown
type DashboardInvestorBreakdown struct {
	InvestorID   int64  `json:"investor_id"`
	InvestorName string `json:"investor_name"`
	PercentPct   string `json:"percent_pct"`
}

// DashboardInvestorContributions representa las contribuciones de inversores
type DashboardInvestorContributions struct {
	ProgressPct decimal.Decimal              `json:"progress_pct"`
	Breakdown   []DashboardInvestorBreakdown `json:"breakdown"`
}

// DashboardOperatingResult representa el resultado operativo
type DashboardOperatingResult struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	IncomeUSD     decimal.Decimal `json:"income_usd"`
	TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
}

// DashboardMetrics representa las métricas principales del dashboard
type DashboardMetrics struct {
	Sowing                *DashboardSowing                `json:"sowing"`
	Harvest               *DashboardHarvest               `json:"harvest"`
	Costs                 *DashboardCosts                 `json:"costs"`
	InvestorContributions *DashboardInvestorContributions `json:"investor_contributions"`
	OperatingResult       *DashboardOperatingResult       `json:"operating_result"`
}

// DashboardBalanceSummary representa el resumen del balance de gestión
type DashboardBalanceSummary struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	StockUSD               decimal.Decimal `json:"stock_usd"`
	RentUSD                decimal.Decimal `json:"rent_usd"`
	StructureUSD           decimal.Decimal `json:"structure_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
}

// DashboardBalanceBreakdown representa un elemento del breakdown del balance
type DashboardBalanceBreakdown struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
}

// DashboardBalanceTotals representa la fila de totales del balance
type DashboardBalanceTotals struct {
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	InvestedUSD decimal.Decimal `json:"invested_usd"`
	StockUSD    decimal.Decimal `json:"stock_usd"`
}

// DashboardManagementBalance representa el balance de gestión completo
type DashboardManagementBalance struct {
	Summary   *DashboardBalanceSummary    `json:"summary"`
	Breakdown []DashboardBalanceBreakdown `json:"breakdown"`
	TotalsRow *DashboardBalanceTotals     `json:"totals_row"`
}

// DashboardCrop representa un cultivo individual
type DashboardCrop struct {
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	RotationPct  decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// DashboardCropTotal representa los totales de cultivos
type DashboardCropTotal struct {
	Hectares          decimal.Decimal `json:"hectares"`
	RotationPct       decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
}

// DashboardCropIncidence representa la incidencia de cultivos completa
type DashboardCropIncidence struct {
	Crops []DashboardCrop     `json:"crops"`
	Total *DashboardCropTotal `json:"total"`
}

// DashboardOperationalCard representa una tarjeta de indicador operativo
type DashboardOperationalCard struct {
	Key           string      `json:"key"`
	Title         string      `json:"title"`
	Date          *string     `json:"date"`
	WorkorderID   interface{} `json:"workorder_id"`
	WorkorderCode interface{} `json:"workorder_code"`
	AuditID       interface{} `json:"audit_id"`
	AuditCode     interface{} `json:"audit_code"`
	Status        interface{} `json:"status"`
}

// DashboardOperationalIndicators representa los indicadores operativos
type DashboardOperationalIndicators struct {
	Cards []DashboardOperationalCard `json:"cards"`
}

// Dashboard representa la entidad principal del dashboard
type Dashboard struct {
	Metrics               *DashboardMetrics               `json:"metrics"`
	ManagementBalance     *DashboardManagementBalance     `json:"management_balance"`
	CropIncidence         *DashboardCropIncidence         `json:"crop_incidence"`
	OperationalIndicators *DashboardOperationalIndicators `json:"operational_indicators"`
}

// ===== ENTIDADES EXISTENTES (MANTENER COMPATIBILIDAD) =====

// ===== Grupo 1: LAS 5 CARDS =====
type DashboardCards struct {
	ProjectID  int64
	CustomerID int64
	CampaignID int64
	FieldID    int64

	// Métricas base
	TotalHectares decimal.Decimal
	SowedArea     decimal.Decimal
	HarvestedArea decimal.Decimal

	// 1) Avance de siembra
	SowingProgressPct decimal.Decimal

	// 2) Avance de costos
	BudgetCostUSD    decimal.Decimal
	CostsProgressPct decimal.Decimal

	// 3) Avance de cosecha
	HarvestProgressPct decimal.Decimal

	// 4) Avance de aportes
	ContributionsProgressPct decimal.Decimal

	// 5) Resultado operativo
	IncomeUSD          decimal.Decimal
	OperatingResultUSD decimal.Decimal
	OperatingResultPct decimal.Decimal

	// Extras útiles
	LaborsExecutedUSD   decimal.Decimal
	SuppliesExecutedUSD decimal.Decimal
	SeedExecutedUSD     decimal.Decimal
	LaborsInvestedUSD   decimal.Decimal
	SuppliesInvestedUSD decimal.Decimal
	SeedInvestedUSD     decimal.Decimal
	StockUSD            decimal.Decimal
	StructureUSD        decimal.Decimal
	RentUSD             decimal.Decimal
}

// ===== Grupo 2: BALANCE DE GESTIÓN =====
type DashboardBalance struct {
	ProjectID  int64
	CustomerID int64
	CampaignID int64
	FieldID    int64

	// Ejecutados
	DirectCostsExecutedUSD decimal.Decimal
	SeedExecutedUSD        decimal.Decimal
	SuppliesExecutedUSD    decimal.Decimal
	LaborsExecutedUSD      decimal.Decimal

	// Invertidos
	DirectCostsInvestedUSD decimal.Decimal
	SeedInvestedUSD        decimal.Decimal
	SuppliesInvestedUSD    decimal.Decimal
	LaborsInvestedUSD      decimal.Decimal

	// Stock y estructura
	StockUSD     decimal.Decimal
	RentUSD      decimal.Decimal
	StructureUSD decimal.Decimal
}

// ===== Grupo 3: INCIDENCIA DE COSTOS POR CULTIVO =====
type DashboardCropIncidenceOld struct {
	ProjectID  int64
	CustomerID int64
	CampaignID int64
	FieldID    int64

	CropName     string
	SurfaceHas   decimal.Decimal
	RotationPct  decimal.Decimal
	TotalCostUSD decimal.Decimal
	CostUSDPerHa decimal.Decimal
	IncidencePct decimal.Decimal
}

// ===== Grupo 4: INDICADORES OPERATIVOS =====
type DashboardOperationalIndicatorsOld struct {
	ProjectID  int64
	CustomerID int64
	CampaignID int64

	FirstWorkorderDate *time.Time
	FirstWorkorderID   *int64
	LastWorkorderDate  *time.Time
	LastWorkorderID    *int64
	LastStockAuditDate *time.Time
	CampaignCloseDate  *time.Time
}

// ===== RESPUESTA COMPLETA DEL DASHBOARD =====
type DashboardResponse struct {
	// Grupo 1: Las 5 cards
	Cards []DashboardCards

	// Grupo 2: Balance de gestión
	Balance []DashboardBalance

	// Grupo 3: Incidencia por cultivo
	CropIncidence []DashboardCropIncidenceOld

	// Grupo 4: Indicadores operativos
	OperationalIndicators []DashboardOperationalIndicatorsOld
}
