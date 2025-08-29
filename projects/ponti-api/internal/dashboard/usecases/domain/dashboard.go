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
type DashboardCropIncidence struct {
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
type DashboardOperationalIndicators struct {
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
	CropIncidence []DashboardCropIncidence

	// Grupo 4: Indicadores operativos
	OperationalIndicators []DashboardOperationalIndicators
}
