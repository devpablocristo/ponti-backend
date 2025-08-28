package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type DashboardFilter struct {
	CampaignID *int64
	ProjectID  *int64
	CustomerID *int64
	FieldID    *int64
	Limit      int
	Offset     int
}

type DashboardRow struct {
	// ===== Grupo 1: MÉTRICAS PRINCIPALES =====
	TotalHectares      decimal.Decimal
	SowedArea          decimal.Decimal
	HarvestedArea      decimal.Decimal
	SowingProgressPct  decimal.NullDecimal
	HarvestProgressPct decimal.NullDecimal

	// VALORES DESCOMPUESTOS PARA AVANCE DE SIEMBRA
	SowingHectares         decimal.Decimal
	TotalHectaresForSowing decimal.Decimal

	// VALORES DESCOMPUESTOS PARA AVANCE DE COSECHA
	HarvestHectares         decimal.Decimal
	TotalHectaresForHarvest decimal.Decimal

	// VALORES DESCOMPUESTOS PARA AVANCE DE COSTOS
	ExecutedCosts decimal.Decimal
	BudgetCosts   decimal.NullDecimal

	// VALORES DESCOMPUESTOS PARA RESULTADO OPERATIVO
	IncomeNet  decimal.Decimal
	TotalCosts decimal.Decimal

	// VALORES DESCOMPUESTOS PARA APORTES DE INVERSORES
	ContributionDetails *string

	LaborsCostUSD      decimal.Decimal
	InputsCostUSD      decimal.Decimal
	ExecutedCostUSD    decimal.Decimal
	BudgetCostUSD      decimal.NullDecimal
	CostsProgressPct   decimal.NullDecimal
	IncomeNetTotalUSD  decimal.Decimal
	AdminTotalUSD      decimal.Decimal
	RentTotalUSD       decimal.Decimal
	OperatingResultUSD decimal.Decimal
	OperatingResultPct decimal.NullDecimal
	InvestedCostUSD    decimal.NullDecimal
	StockUSD           decimal.NullDecimal

	// ===== Grupo 2: APORTES DE INVERSORES =====
	InvestorContributionPct decimal.NullDecimal
	ContributionBreakdown   *string

	// ===== Grupo 3: INCIDENCIA POR CULTIVO =====
	CropsBreakdown           *string
	CropsDetails             *string
	CropsTotalHectares       decimal.Decimal
	CropsTotalRotationPct    decimal.Decimal
	CropsTotalCostPerHectare decimal.Decimal

	// ===== Grupo 4: RENDIMIENTO Y COSTOS =====
	YieldPerHectare     decimal.NullDecimal
	TotalCostPerHectare decimal.NullDecimal

	// ===== Grupo 5: INDICADORES OPERATIVOS =====
	FirstOrderDate     *time.Time
	FirstOrderNumber   *string
	LastOrderDate      *time.Time
	LastOrderNumber    *string
	LastStockCountDate *time.Time

	// ===== CAMPOS ADICIONALES PARA COMPATIBILIDAD =====
	MgmtIncomeUSD          decimal.Decimal
	MgmtTotalCostsUSD      decimal.Decimal
	MgmtOperatingResultUSD decimal.Decimal
	MgmtOperatingResultPct decimal.NullDecimal

	// ===== BALANCE DE GESTIÓN DETALLADO =====
	// Direct costs
	DirectCostsExecutedUSD decimal.Decimal
	DirectCostsInvestedUSD decimal.Decimal
	DirectCostsStockUSD    decimal.Decimal
	DirectCostsHectares    decimal.Decimal

	// Seed
	SeedExecutedUSD decimal.Decimal
	SeedInvestedUSD decimal.Decimal
	SeedStockUSD    decimal.Decimal
	SeedHectares    decimal.Decimal

	// Supplies
	SuppliesExecutedUSD decimal.Decimal
	SuppliesInvestedUSD decimal.Decimal
	SuppliesStockUSD    decimal.Decimal
	SuppliesHectares    decimal.Decimal

	// Labors
	LaborsExecutedUSD decimal.Decimal
	LaborsInvestedUSD decimal.Decimal
	LaborsStockUSD    decimal.Decimal
	LaborsHectares    decimal.Decimal

	// Rent
	RentExecutedUSD decimal.Decimal
	RentInvestedUSD decimal.Decimal
	RentStockUSD    decimal.Decimal
	RentHectares    decimal.Decimal

	// Structure
	StructureExecutedUSD decimal.Decimal
	StructureInvestedUSD decimal.Decimal
	StructureStockUSD    decimal.Decimal
	StructureHectares    decimal.Decimal
}
