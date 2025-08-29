package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ===== Grupo 1: LAS 5 CARDS =====
type DashboardCards struct {
	ProjectID  int64 `gorm:"column:project_id"`
	CustomerID int64 `gorm:"column:customer_id"`
	CampaignID int64 `gorm:"column:campaign_id"`
	FieldID    int64 `gorm:"column:field_id"`

	// Métricas base
	TotalHectares decimal.Decimal `gorm:"column:total_hectares"`
	SowedArea     decimal.Decimal `gorm:"column:sowed_area"`
	HarvestedArea decimal.Decimal `gorm:"column:harvested_area"`

	// 1) Avance de siembra
	SowingProgressPct decimal.Decimal `gorm:"column:sowing_progress_pct"`

	// 2) Avance de costos
	BudgetCostUSD    decimal.Decimal `gorm:"column:budget_cost_usd"`
	CostsProgressPct decimal.Decimal `gorm:"column:costs_progress_pct"`

	// 3) Avance de cosecha
	HarvestProgressPct decimal.Decimal `gorm:"column:harvest_progress_pct"`

	// 4) Avance de aportes
	ContributionsProgressPct decimal.Decimal `gorm:"column:contributions_progress_pct"`

	// 5) Resultado operativo
	IncomeUSD          decimal.Decimal `gorm:"column:income_usd"`
	OperatingResultUSD decimal.Decimal `gorm:"column:operating_result_usd"`
	OperatingResultPct decimal.Decimal `gorm:"column:operating_result_pct"`

	// Extras útiles
	LaborsExecutedUSD   decimal.Decimal `gorm:"column:labors_executed_usd"`
	SuppliesExecutedUSD decimal.Decimal `gorm:"column:supplies_executed_usd"`
	SeedExecutedUSD     decimal.Decimal `gorm:"column:seed_executed_usd"`
	LaborsInvestedUSD   decimal.Decimal `gorm:"column:labors_invested_usd"`
	SuppliesInvestedUSD decimal.Decimal `gorm:"column:supplies_invested_usd"`
	SeedInvestedUSD     decimal.Decimal `gorm:"column:seed_invested_usd"`
	StockUSD            decimal.Decimal `gorm:"column:stock_usd"`
	StructureUSD        decimal.Decimal `gorm:"column:structure_usd"`
	RentUSD             decimal.Decimal `gorm:"column:rent_usd"`
}

// ===== Grupo 2: BALANCE DE GESTIÓN =====
type DashboardBalance struct {
	ProjectID  int64 `gorm:"column:project_id"`
	CustomerID int64 `gorm:"column:customer_id"`
	CampaignID int64 `gorm:"column:campaign_id"`
	FieldID    int64 `gorm:"column:field_id"`

	// Ejecutados
	DirectCostsExecutedUSD decimal.Decimal `gorm:"column:direct_costs_executed_usd"`
	SeedExecutedUSD        decimal.Decimal `gorm:"column:seed_executed_usd"`
	SuppliesExecutedUSD    decimal.Decimal `gorm:"column:supplies_executed_usd"`
	LaborsExecutedUSD      decimal.Decimal `gorm:"column:labors_executed_usd"`

	// Invertidos
	DirectCostsInvestedUSD decimal.Decimal `gorm:"column:direct_costs_invested_usd"`
	SeedInvestedUSD        decimal.Decimal `gorm:"column:seed_invested_usd"`
	SuppliesInvestedUSD    decimal.Decimal `gorm:"column:supplies_invested_usd"`
	LaborsInvestedUSD      decimal.Decimal `gorm:"column:labors_invested_usd"`

	// Stock y estructura
	StockUSD     decimal.Decimal `gorm:"column:stock_usd"`
	RentUSD      decimal.Decimal `gorm:"column:rent_usd"`
	StructureUSD decimal.Decimal `gorm:"column:structure_usd"`
}

// ===== Grupo 3: INCIDENCIA DE COSTOS POR CULTIVO =====
type DashboardCropIncidence struct {
	ProjectID  int64 `gorm:"column:project_id"`
	CustomerID int64 `gorm:"column:customer_id"`
	CampaignID int64 `gorm:"column:campaign_id"`
	FieldID    int64 `gorm:"column:field_id"`

	CropName     string          `gorm:"column:crop_name"`
	SurfaceHas   decimal.Decimal `gorm:"column:surface_has"`
	RotationPct  decimal.Decimal `gorm:"column:rotation_pct"`
	TotalCostUSD decimal.Decimal `gorm:"column:total_cost_usd"`
	CostUSDPerHa decimal.Decimal `gorm:"column:cost_usd_per_ha"`
	IncidencePct decimal.Decimal `gorm:"column:incidence_pct"`
}

// ===== Grupo 4: INDICADORES OPERATIVOS =====
type DashboardOperationalIndicators struct {
	ProjectID  int64 `gorm:"column:project_id"`
	CustomerID int64 `gorm:"column:customer_id"`
	CampaignID int64 `gorm:"column:campaign_id"`

	FirstWorkorderDate *time.Time `gorm:"column:first_workorder_date"`
	FirstWorkorderID   *int64     `gorm:"column:first_workorder_id"`
	LastWorkorderDate  *time.Time `gorm:"column:last_workorder_date"`
	LastWorkorderID    *int64     `gorm:"column:last_workorder_id"`
	LastStockAuditDate *time.Time `gorm:"column:last_stock_audit_date"`
	CampaignCloseDate  *time.Time `gorm:"column:campaign_close_date"`
}

// ===== MODELO LEGACY (mantener para compatibilidad) =====
type DashboardRow struct {
	// Basic metrics
	TotalHectares      decimal.Decimal     `gorm:"column:total_hectares"`
	SowedArea          decimal.Decimal     `gorm:"column:sowed_area"`
	HarvestedArea      decimal.Decimal     `gorm:"column:harvested_area"`
	SowingProgressPct  decimal.NullDecimal `gorm:"column:sowing_progress_pct"`
	HarvestProgressPct decimal.NullDecimal `gorm:"column:harvest_progress_pct"`

	// Labor costs
	LaborsExecutedUSD      decimal.Decimal `gorm:"column:labors_executed_usd"`
	SuppliesExecutedUSD    decimal.Decimal `gorm:"column:supplies_executed_usd"`
	SeedExecutedUSD        decimal.Decimal `gorm:"column:seed_executed_usd"`
	DirectCostsExecutedUSD decimal.Decimal `gorm:"column:direct_costs_executed_usd"`

	// Project costs
	LaborsInvestedUSD      decimal.Decimal `gorm:"column:labors_invested_usd"`
	SuppliesInvestedUSD    decimal.Decimal `gorm:"column:supplies_invested_usd"`
	DirectCostsInvestedUSD decimal.Decimal `gorm:"column:direct_costs_invested_usd"`

	// Stock and budget
	StockUSD         decimal.Decimal     `gorm:"column:stock_usd"`
	BudgetCostUSD    decimal.Decimal     `gorm:"column:budget_cost_usd"`
	CostsProgressPct decimal.NullDecimal `gorm:"column:costs_progress_pct"`

	// Income and structure
	IncomeUSD    decimal.Decimal `gorm:"column:income_usd"`
	RentUSD      decimal.Decimal `gorm:"column:rent_usd"`
	StructureUSD decimal.Decimal `gorm:"column:structure_usd"`

	// Operating result
	OperatingResultUSD decimal.Decimal     `gorm:"column:operating_result_usd"`
	OperatingResultPct decimal.NullDecimal `gorm:"column:operating_result_pct"`

	// Cost per hectare
	TotalCostPerHectare decimal.Decimal `gorm:"column:total_cost_per_hectare"`

	// Crops breakdown
	CropsBreakdown *string `gorm:"column:crops_breakdown"`

	// Additional fields needed for DTO mapping
	SeedInvestedUSD decimal.Decimal `gorm:"column:seed_invested_usd"`
}

// ===== MODELO PARA LA FUNCIÓN SQL get_dashboard_payload =====

// DashboardPayloadResponse representa la respuesta directa de la función SQL get_dashboard_payload
type DashboardPayloadResponse struct {
	Metrics               DashboardPayloadMetrics               `json:"metrics"`
	ManagementBalance     DashboardPayloadManagementBalance     `json:"management_balance"`
	CropIncidence         DashboardPayloadCropIncidence         `json:"crop_incidence"`
	OperationalIndicators DashboardPayloadOperationalIndicators `json:"operational_indicators"`
}

// DashboardPayloadMetrics representa las métricas del dashboard
type DashboardPayloadMetrics struct {
	Sowing                DashboardPayloadSowingMetric          `json:"sowing"`
	Harvest               DashboardPayloadHarvestMetric         `json:"harvest"`
	Costs                 DashboardPayloadCostsMetric           `json:"costs"`
	InvestorContributions DashboardPayloadInvestorContributions `json:"investor_contributions"`
	OperatingResult       DashboardPayloadOperatingResultMetric `json:"operating_result"`
}

// DashboardPayloadSowingMetric representa la métrica de siembra
type DashboardPayloadSowingMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// DashboardPayloadHarvestMetric representa la métrica de cosecha
type DashboardPayloadHarvestMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// DashboardPayloadCostsMetric representa la métrica de costos
type DashboardPayloadCostsMetric struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	BudgetUSD   decimal.Decimal `json:"budget_usd"`
}

// DashboardPayloadInvestorContributions representa la métrica de contribuciones de inversores
type DashboardPayloadInvestorContributions struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	Breakdown   interface{}     `json:"breakdown"`
}

// DashboardPayloadOperatingResultMetric representa la métrica de resultado operativo
type DashboardPayloadOperatingResultMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	IncomeUSD     decimal.Decimal `json:"income_usd"`
	TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
}

// DashboardPayloadManagementBalance representa el balance de gestión
type DashboardPayloadManagementBalance struct {
	Summary   DashboardPayloadBalanceSummary     `json:"summary"`
	Breakdown []DashboardPayloadBalanceBreakdown `json:"breakdown"`
	TotalsRow DashboardPayloadBalanceTotals      `json:"totals_row"`
}

// DashboardPayloadBalanceSummary representa el resumen del balance
type DashboardPayloadBalanceSummary struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	StockUSD               decimal.Decimal `json:"stock_usd"`
	RentUSD                decimal.Decimal `json:"rent_usd"`
	StructureUSD           decimal.Decimal `json:"structure_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
}

// DashboardPayloadBalanceBreakdown representa el desglose por categoría
type DashboardPayloadBalanceBreakdown struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
}

// DashboardPayloadBalanceTotals representa la fila de totales
type DashboardPayloadBalanceTotals struct {
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	InvestedUSD decimal.Decimal `json:"invested_usd"`
	StockUSD    decimal.Decimal `json:"stock_usd"`
}

// DashboardPayloadCropIncidence representa la incidencia por cultivo
type DashboardPayloadCropIncidence struct {
	Crops []DashboardPayloadCropData `json:"crops"`
	Total DashboardPayloadCropTotal  `json:"total"`
}

// DashboardPayloadCropData representa los datos de un cultivo específico
type DashboardPayloadCropData struct {
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	RotationPct  decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// DashboardPayloadCropTotal representa los totales de cultivos
type DashboardPayloadCropTotal struct {
	Hectares          decimal.Decimal `json:"hectares"`
	RotationPct       decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
}

// DashboardPayloadOperationalIndicators representa los indicadores operativos
type DashboardPayloadOperationalIndicators struct {
	Cards []DashboardPayloadOperationalCard `json:"cards"`
}

// DashboardPayloadOperationalCard representa una tarjeta de indicador operativo
type DashboardPayloadOperationalCard struct {
	Key           string     `json:"key"`
	Title         string     `json:"title"`
	Date          *time.Time `json:"date"`
	WorkorderID   *int64     `json:"workorder_id"`
	WorkorderCode *string    `json:"workorder_code"`
	AuditID       *int64     `json:"audit_id"`
	AuditCode     *string    `json:"audit_code"`
	Status        *string    `json:"status"`
}
