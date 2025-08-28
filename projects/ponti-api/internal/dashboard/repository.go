package dashboard

import (
	"context"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	m "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetDashboard(ctx context.Context, filt domain.DashboardFilter) (*domain.DashboardRow, error) {
	q := r.db.Client().WithContext(ctx).
		Table("dashboard_full_view").
		Select([]string{
			// ===== Grupo 1: MÉTRICAS PRINCIPALES =====
			"total_hectares", "sowed_area", "harvested_area",
			"sowing_progress_pct", "harvest_progress_pct",

			// VALORES DESCOMPUESTOS PARA AVANCE DE SIEMBRA
			"sowing_hectares", "total_hectares_for_sowing",

			// VALORES DESCOMPUESTOS PARA AVANCE DE COSECHA
			"harvest_hectares", "total_hectares_for_harvest",

			// VALORES DESCOMPUESTOS PARA AVANCE DE COSTOS
			"executed_costs", "budget_costs",

			// VALORES DESCOMPUESTOS PARA RESULTADO OPERATIVO
			"income_net", "total_costs",

			// VALORES DESCOMPUESTOS PARA APORTES DE INVERSORES
			"contribution_details",

			"labors_cost_usd", "inputs_cost_usd",
			"executed_cost_usd", "budget_cost_usd", "costs_progress_pct",
			"income_net_total_usd", "admin_total_usd", "rent_total_usd",
			"operating_result_usd", "operating_result_pct",
			"invested_cost_usd", "stock_usd",

			// ===== Grupo 2: APORTES DE INVERSORES =====
			"investor_contribution_pct", "contribution_breakdown",

			// ===== Grupo 3: INCIDENCIA POR CULTIVO =====
			"crops_breakdown", "crops_details", "crops_total_hectares",
			"crops_total_rotation_pct", "crops_total_cost_per_hectare",

			// ===== Grupo 4: RENDIMIENTO Y COSTOS =====
			"yield_per_hectare", "total_cost_per_hectare",

			// ===== Grupo 5: INDICADORES OPERATIVOS =====
			"first_order_date", "first_order_number", "last_order_date", "last_order_number",
			"last_stock_count_date",

			// ===== CAMPOS ADICIONALES PARA COMPATIBILIDAD =====
			"mgmt_income_usd", "mgmt_total_costs_usd", "mgmt_operating_result_usd", "mgmt_operating_result_pct",

			// ===== BALANCE DE GESTIÓN DETALLADO =====
			"direct_costs_executed_usd", "direct_costs_invested_usd", "direct_costs_stock_usd", "direct_costs_hectares",
			"seed_executed_usd", "seed_invested_usd", "seed_stock_usd", "seed_hectares",
			"supplies_executed_usd", "supplies_invested_usd", "supplies_stock_usd", "supplies_hectares",
			"labors_executed_usd", "labors_invested_usd", "labors_stock_usd", "labors_hectares",
			"rent_executed_usd", "rent_invested_usd", "rent_stock_usd", "rent_hectares",
			"structure_executed_usd", "structure_invested_usd", "structure_stock_usd", "structure_hectares",
		})

	if filt.CampaignID != nil {
		q = q.Where("campaign_id = ?", *filt.CampaignID)
	}
	if filt.ProjectID != nil {
		q = q.Where("project_id  = ?", *filt.ProjectID)
	}
	if filt.CustomerID != nil {
		q = q.Where("customer_id = ?", *filt.CustomerID)
	}
	if filt.FieldID != nil {
		q = q.Where("field_id = ?", *filt.FieldID)
	}
	q = q.Order("campaign_id, project_id, customer_id")
	q = q.Limit(1) // SOLO UN REGISTRO

	var row m.DashboardRow
	if err := q.Scan(&row).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard", err)
	}

	result := &domain.DashboardRow{
		// ===== Grupo 1: MÉTRICAS PRINCIPALES =====
		TotalHectares:      row.TotalHectares,
		SowedArea:          row.SowedArea,
		HarvestedArea:      row.HarvestedArea,
		SowingProgressPct:  row.SowingProgressPct,
		HarvestProgressPct: row.HarvestProgressPct,

		// VALORES DESCOMPUESTOS PARA AVANCE DE SIEMBRA
		SowingHectares:         row.SowingHectares,
		TotalHectaresForSowing: row.TotalHectaresForSowing,

		// VALORES DESCOMPUESTOS PARA AVANCE DE COSECHA
		HarvestHectares:         row.HarvestHectares,
		TotalHectaresForHarvest: row.TotalHectaresForHarvest,

		// VALORES DESCOMPUESTOS PARA AVANCE DE COSTOS
		ExecutedCosts: row.ExecutedCosts,
		BudgetCosts:   row.BudgetCosts,

		// VALORES DESCOMPUESTOS PARA RESULTADO OPERATIVO
		IncomeNet:  row.IncomeNet,
		TotalCosts: row.TotalCosts,

		// VALORES DESCOMPUESTOS PARA APORTES DE INVERSORES
		ContributionDetails: row.ContributionDetails,

		LaborsCostUSD:      row.LaborsCostUSD,
		InputsCostUSD:      row.InputsCostUSD,
		ExecutedCostUSD:    row.ExecutedCostUSD,
		BudgetCostUSD:      row.BudgetCostUSD,
		CostsProgressPct:   row.CostsProgressPct,
		IncomeNetTotalUSD:  row.IncomeNetTotalUSD,
		AdminTotalUSD:      row.AdminTotalUSD,
		RentTotalUSD:       row.RentTotalUSD,
		OperatingResultUSD: row.OperatingResultUSD,
		OperatingResultPct: row.OperatingResultPct,
		InvestedCostUSD:    row.InvestedCostUSD,
		StockUSD:           row.StockUSD,

		// ===== Grupo 2: APORTES DE INVERSORES =====
		InvestorContributionPct: row.InvestorContributionPct,
		ContributionBreakdown:   row.ContributionBreakdown,

		// ===== Grupo 3: INCIDENCIA POR CULTIVO =====
		CropsBreakdown:           row.CropsBreakdown,
		CropsDetails:             row.CropsDetails,
		CropsTotalHectares:       row.CropsTotalHectares,
		CropsTotalRotationPct:    row.CropsTotalRotationPct,
		CropsTotalCostPerHectare: row.CropsTotalCostPerHectare,

		// ===== Grupo 4: RENDIMIENTO Y COSTOS =====
		YieldPerHectare:     row.YieldPerHectare,
		TotalCostPerHectare: row.TotalCostPerHectare,

		// ===== Grupo 5: INDICADORES OPERATIVOS =====
		FirstOrderDate:     row.FirstOrderDate,
		FirstOrderNumber:   row.FirstOrderNumber,
		LastOrderDate:      row.LastOrderDate,
		LastOrderNumber:    row.LastOrderNumber,
		LastStockCountDate: row.LastStockCountDate,

		// ===== CAMPOS ADICIONALES PARA COMPATIBILIDAD =====
		MgmtIncomeUSD:          row.MgmtIncomeUSD,
		MgmtTotalCostsUSD:      row.MgmtTotalCostsUSD,
		MgmtOperatingResultUSD: row.MgmtOperatingResultUSD,
		MgmtOperatingResultPct: row.MgmtOperatingResultPct,

		// ===== BALANCE DE GESTIÓN DETALLADO =====
		// Direct costs
		DirectCostsExecutedUSD: row.DirectCostsExecutedUSD,
		DirectCostsInvestedUSD: row.DirectCostsInvestedUSD,
		DirectCostsStockUSD:    row.DirectCostsStockUSD,
		DirectCostsHectares:    row.DirectCostsHectares,

		// Seed
		SeedExecutedUSD: row.SeedExecutedUSD,
		SeedInvestedUSD: row.SeedInvestedUSD,
		SeedStockUSD:    row.SeedStockUSD,
		SeedHectares:    row.SeedHectares,

		// Supplies
		SuppliesExecutedUSD: row.SuppliesExecutedUSD,
		SuppliesInvestedUSD: row.SuppliesInvestedUSD,
		SuppliesStockUSD:    row.SuppliesStockUSD,
		SuppliesHectares:    row.SuppliesHectares,

		// Labors
		LaborsExecutedUSD: row.LaborsExecutedUSD,
		LaborsInvestedUSD: row.LaborsInvestedUSD,
		LaborsStockUSD:    row.LaborsStockUSD,
		LaborsHectares:    row.LaborsHectares,

		// Rent
		RentExecutedUSD: row.RentExecutedUSD,
		RentInvestedUSD: row.RentInvestedUSD,
		RentStockUSD:    row.RentStockUSD,
		RentHectares:    row.RentHectares,

		// Structure
		StructureExecutedUSD: row.StructureExecutedUSD,
		StructureInvestedUSD: row.StructureInvestedUSD,
		StructureStockUSD:    row.StructureStockUSD,
		StructureHectares:    row.StructureHectares,
	}

	return result, nil
}
