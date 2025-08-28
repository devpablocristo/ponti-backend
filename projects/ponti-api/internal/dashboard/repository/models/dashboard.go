package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// DashboardRow maps the SQL view dashboard_full_view for scanning via GORM.Raw.
type DashboardRow struct {
	// ===== Grupo 1: MÉTRICAS PRINCIPALES =====
	TotalHectares      decimal.Decimal     `gorm:"column:total_hectares"`
	SowedArea          decimal.Decimal     `gorm:"column:sowed_area"`
	HarvestedArea      decimal.Decimal     `gorm:"column:harvested_area"`
	SowingProgressPct  decimal.NullDecimal `gorm:"column:sowing_progress_pct"`
	HarvestProgressPct decimal.NullDecimal `gorm:"column:harvest_progress_pct"`

	// VALORES DESCOMPUESTOS PARA AVANCE DE SIEMBRA
	SowingHectares         decimal.Decimal `gorm:"column:sowing_hectares"`
	TotalHectaresForSowing decimal.Decimal `gorm:"column:total_hectares_for_sowing"`

	// VALORES DESCOMPUESTOS PARA AVANCE DE COSECHA
	HarvestHectares         decimal.Decimal `gorm:"column:harvest_hectares"`
	TotalHectaresForHarvest decimal.Decimal `gorm:"column:total_hectares_for_harvest"`

	// VALORES DESCOMPUESTOS PARA AVANCE DE COSTOS
	ExecutedCosts decimal.Decimal     `gorm:"column:executed_costs"`
	BudgetCosts   decimal.NullDecimal `gorm:"column:budget_costs"`

	// VALORES DESCOMPUESTOS PARA RESULTADO OPERATIVO
	IncomeNet  decimal.Decimal `gorm:"column:income_net"`
	TotalCosts decimal.Decimal `gorm:"column:total_costs"`

	// VALORES DESCOMPUESTOS PARA APORTES DE INVERSORES
	ContributionDetails *string `gorm:"column:contribution_details"`

	LaborsCostUSD      decimal.Decimal     `gorm:"column:labors_cost_usd"`
	InputsCostUSD      decimal.Decimal     `gorm:"column:inputs_cost_usd"`
	ExecutedCostUSD    decimal.Decimal     `gorm:"column:executed_cost_usd"`
	BudgetCostUSD      decimal.NullDecimal `gorm:"column:budget_cost_usd"`
	CostsProgressPct   decimal.NullDecimal `gorm:"column:costs_progress_pct"`
	IncomeNetTotalUSD  decimal.Decimal     `gorm:"column:income_net_total_usd"`
	AdminTotalUSD      decimal.Decimal     `gorm:"column:admin_total_usd"`
	RentTotalUSD       decimal.Decimal     `gorm:"column:rent_total_usd"`
	OperatingResultUSD decimal.Decimal     `gorm:"column:operating_result_usd"`
	OperatingResultPct decimal.NullDecimal `gorm:"column:operating_result_pct"`
	InvestedCostUSD    decimal.NullDecimal `gorm:"column:invested_cost_usd"`
	StockUSD           decimal.NullDecimal `gorm:"column:stock_usd"`

	// ===== Grupo 2: APORTES DE INVERSORES =====
	InvestorContributionPct decimal.NullDecimal `gorm:"column:investor_contribution_pct"`
	ContributionBreakdown   *string             `gorm:"column:contribution_breakdown"`

	// ===== Grupo 3: INCIDENCIA POR CULTIVO =====
	CropsBreakdown           *string         `gorm:"column:crops_breakdown"`
	CropsDetails             *string         `gorm:"column:crops_details"`
	CropsTotalHectares       decimal.Decimal `gorm:"column:crops_total_hectares"`
	CropsTotalRotationPct    decimal.Decimal `gorm:"column:crops_total_rotation_pct"`
	CropsTotalCostPerHectare decimal.Decimal `gorm:"column:crops_total_cost_per_hectare"`

	// ===== Grupo 4: RENDIMIENTO Y COSTOS =====
	YieldPerHectare     decimal.NullDecimal `gorm:"column:yield_per_hectare"`
	TotalCostPerHectare decimal.NullDecimal `gorm:"column:total_cost_per_hectare"`

	// ===== Grupo 5: INDICADORES OPERATIVOS =====
	FirstOrderDate     *time.Time `gorm:"column:first_order_date"`
	FirstOrderNumber   *string    `gorm:"column:first_order_number"`
	LastOrderDate      *time.Time `gorm:"column:last_order_date"`
	LastOrderNumber    *string    `gorm:"column:last_order_number"`
	LastStockCountDate *time.Time `gorm:"column:last_stock_count_date"`

	// ===== CAMPOS ADICIONALES PARA COMPATIBILIDAD =====
	MgmtIncomeUSD          decimal.Decimal     `gorm:"column:mgmt_income_usd"`
	MgmtTotalCostsUSD      decimal.Decimal     `gorm:"column:mgmt_total_costs_usd"`
	MgmtOperatingResultUSD decimal.Decimal     `gorm:"column:mgmt_operating_result_usd"`
	MgmtOperatingResultPct decimal.NullDecimal `gorm:"column:mgmt_operating_result_pct"`

	// ===== BALANCE DE GESTIÓN DETALLADO =====
	// Direct costs
	DirectCostsExecutedUSD decimal.Decimal `gorm:"column:direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `gorm:"column:direct_costs_invested_usd"`
	DirectCostsStockUSD    decimal.Decimal `gorm:"column:direct_costs_stock_usd"`
	DirectCostsHectares    decimal.Decimal `gorm:"column:direct_costs_hectares"`

	// Seed
	SeedExecutedUSD decimal.Decimal `gorm:"column:seed_executed_usd"`
	SeedInvestedUSD decimal.Decimal `gorm:"column:seed_invested_usd"`
	SeedStockUSD    decimal.Decimal `gorm:"column:seed_stock_usd"`
	SeedHectares    decimal.Decimal `gorm:"column:seed_hectares"`

	// Supplies
	SuppliesExecutedUSD decimal.Decimal `gorm:"column:supplies_executed_usd"`
	SuppliesInvestedUSD decimal.Decimal `gorm:"column:supplies_invested_usd"`
	SuppliesStockUSD    decimal.Decimal `gorm:"column:supplies_stock_usd"`
	SuppliesHectares    decimal.Decimal `gorm:"column:supplies_hectares"`

	// Labors
	LaborsExecutedUSD decimal.Decimal `gorm:"column:labors_executed_usd"`
	LaborsInvestedUSD decimal.Decimal `gorm:"column:labors_invested_usd"`
	LaborsStockUSD    decimal.Decimal `gorm:"column:labors_stock_usd"`
	LaborsHectares    decimal.Decimal `gorm:"column:labors_hectares"`

	// Rent
	RentExecutedUSD decimal.Decimal `gorm:"column:rent_executed_usd"`
	RentInvestedUSD decimal.Decimal `gorm:"column:rent_invested_usd"`
	RentStockUSD    decimal.Decimal `gorm:"column:rent_stock_usd"`
	RentHectares    decimal.Decimal `gorm:"column:rent_hectares"`

	// Structure
	StructureExecutedUSD decimal.Decimal `gorm:"column:structure_executed_usd"`
	StructureInvestedUSD decimal.Decimal `gorm:"column:structure_invested_usd"`
	StructureStockUSD    decimal.Decimal `gorm:"column:structure_stock_usd"`
	StructureHectares    decimal.Decimal `gorm:"column:structure_hectares"`
}
