package domain

import (
	"github.com/shopspring/decimal"
)

type DashboardFilter struct {
	CampaignIDs []int64
	ProjectIDs  []int64
	CustomerIDs []int64
	FieldIDs    []int64
	Limit       int
	Offset      int
}

type DashboardRow struct {
	// Basic metrics
	TotalHectares      decimal.Decimal
	SowedArea          decimal.Decimal
	HarvestedArea      decimal.Decimal
	SowingProgressPct  decimal.NullDecimal
	HarvestProgressPct decimal.NullDecimal

	// Labor costs
	LaborsExecutedUSD      decimal.Decimal
	SuppliesExecutedUSD    decimal.Decimal
	SeedExecutedUSD        decimal.Decimal
	DirectCostsExecutedUSD decimal.Decimal

	// Project costs
	LaborsInvestedUSD      decimal.Decimal
	SuppliesInvestedUSD    decimal.Decimal
	DirectCostsInvestedUSD decimal.Decimal

	// Stock and budget
	StockUSD         decimal.Decimal
	BudgetCostUSD    decimal.Decimal
	CostsProgressPct decimal.NullDecimal

	// Income and structure
	IncomeUSD    decimal.Decimal
	RentUSD      decimal.Decimal
	StructureUSD decimal.Decimal

	// Operating result
	OperatingResultUSD decimal.Decimal
	OperatingResultPct decimal.NullDecimal

	// Cost per hectare
	TotalCostPerHectare decimal.Decimal

	// Crops breakdown
	CropsBreakdown *string

	// Additional fields for DTO mapping
	SeedInvestedUSD decimal.Decimal
}
