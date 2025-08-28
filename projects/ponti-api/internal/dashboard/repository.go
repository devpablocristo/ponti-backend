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
			// Basic metrics
			"total_hectares", "sowed_area", "harvested_area",
			"sowing_progress_pct", "harvest_progress_pct",

			// Labor costs
			"labors_executed_usd", "supplies_executed_usd", "seed_executed_usd", "direct_costs_executed_usd",

			// Project costs
			"labors_invested_usd", "supplies_invested_usd", "direct_costs_invested_usd",

			// Stock and budget
			"stock_usd", "budget_cost_usd", "costs_progress_pct",

			// Income and structure
			"income_usd", "rent_usd", "structure_usd",

			// Operating result
			"operating_result_usd", "operating_result_pct",

			// Cost per hectare
			"total_cost_per_hectare",

			// Crops breakdown
			"crops_breakdown",

			// Additional fields for DTO mapping
			"seed_invested_usd",
		})

	// Aplicar filtros de arrays
	if len(filt.CustomerIDs) > 0 {
		q = q.Where("customer_id = ANY(?)", filt.CustomerIDs)
	}
	if len(filt.ProjectIDs) > 0 {
		q = q.Where("project_id = ANY(?)", filt.ProjectIDs)
	}
	if len(filt.CampaignIDs) > 0 {
		q = q.Where("campaign_id = ANY(?)", filt.CampaignIDs)
	}
	if len(filt.FieldIDs) > 0 {
		q = q.Where("field_id = ANY(?)", filt.FieldIDs)
	}

	q = q.Order("campaign_id, project_id, customer_id")
	q = q.Limit(1) // SOLO UN REGISTRO

	var row m.DashboardRow
	if err := q.Scan(&row).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard", err)
	}

	result := &domain.DashboardRow{
		// Basic metrics
		TotalHectares:      row.TotalHectares,
		SowedArea:          row.SowedArea,
		HarvestedArea:      row.HarvestedArea,
		SowingProgressPct:  row.SowingProgressPct,
		HarvestProgressPct: row.HarvestProgressPct,

		// Labor costs
		LaborsExecutedUSD:      row.LaborsExecutedUSD,
		SuppliesExecutedUSD:    row.SuppliesExecutedUSD,
		SeedExecutedUSD:        row.SeedExecutedUSD,
		DirectCostsExecutedUSD: row.DirectCostsExecutedUSD,

		// Project costs
		LaborsInvestedUSD:      row.LaborsInvestedUSD,
		SuppliesInvestedUSD:    row.SuppliesInvestedUSD,
		DirectCostsInvestedUSD: row.DirectCostsInvestedUSD,

		// Stock and budget
		StockUSD:         row.StockUSD,
		BudgetCostUSD:    row.BudgetCostUSD,
		CostsProgressPct: row.CostsProgressPct,

		// Income and structure
		IncomeUSD:    row.IncomeUSD,
		RentUSD:      row.RentUSD,
		StructureUSD: row.StructureUSD,

		// Operating result
		OperatingResultUSD: row.OperatingResultUSD,
		OperatingResultPct: row.OperatingResultPct,

		// Cost per hectare
		TotalCostPerHectare: row.TotalCostPerHectare,

		// Crops breakdown
		CropsBreakdown: row.CropsBreakdown,

		// Additional fields for DTO mapping
		SeedInvestedUSD: row.SeedInvestedUSD,
	}

	return result, nil
}
