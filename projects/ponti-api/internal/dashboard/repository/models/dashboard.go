package models

import (
	"github.com/shopspring/decimal"
)

// DashboardRow maps the SQL view dashboard_full_view for scanning via GORM.Raw.
// Simplified to only include essential columns needed for the dashboard
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
