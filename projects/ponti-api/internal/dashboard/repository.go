package dashboard

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateDashboard(ctx context.Context, d *domain.Dashboard) (int64, error) {
	if d == nil {
		return 0, types.NewError(types.ErrValidation, "dashboard is nil", nil)
	}
	model := models.FromDomainDashboard(d)
	model.Base = sharedmodels.Base{
		CreatedBy: d.CreatedBy,
		UpdatedBy: d.UpdatedBy,
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create dashboard", err)
	}
	return model.ID, nil
}

func (r *Repository) ListDashboards(ctx context.Context) ([]domain.Dashboard, error) {
	var list []models.Dashboard
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list dashboards", err)
	}
	result := make([]domain.Dashboard, 0, len(list))
	for _, d := range list {
		result = append(result, *d.ToDomain())
	}
	return result, nil
}

func (r *Repository) GetDashboard(ctx context.Context, id int64) (*domain.Dashboard, error) {
	var model models.Dashboard
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("dashboard with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) UpdateDashboard(ctx context.Context, d *domain.Dashboard) error {
	if d == nil {
		return types.NewError(types.ErrValidation, "dashboard is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Dashboard{}).
		Where("id = ?", d.ID).
		Updates(models.FromDomainDashboard(d))
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to update dashboard", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("dashboard with id %d does not exist", d.ID), nil)
	}
	return nil
}

func (r *Repository) DeleteDashboard(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Dashboard{}, "id = ?", id)
	if result.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete dashboard", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, fmt.Sprintf("dashboard with id %d does not exist", id), nil)
	}
	return nil
}

func (r *Repository) GetDashboardData(ctx context.Context, filter domain.DashboardFilter) (*domain.DashboardData, error) {
	// Construir la consulta SQL base
	query := `
		SELECT 
			-- Métricas de siembra
			COALESCE(SUM(s.sowed_area), 0) as sowing_hectares,
			COALESCE(SUM(s.total_hectares), 0) as sowing_total_hectares,
			
			-- Métricas de cosecha
			COALESCE(SUM(h.harvested_area), 0) as harvest_hectares,
			COALESCE(SUM(h.total_hectares), 0) as harvest_total_hectares,
			
			-- Métricas de costos
			COALESCE(SUM(ca.executed_costs_usd), 0) as costs_executed_usd,
			COALESCE(SUM(ca.budget_total_usd), 0) as costs_budget_usd,
			COALESCE(SUM(ca.costs_progress_pct), 0) as costs_progress_pct,
			
			-- Resultado operativo
			COALESCE(SUM(o.income_usd), 0) as operating_income_usd,
			COALESCE(SUM(o.total_invested_usd), 0) as operating_total_costs_usd,
			
			-- Balance de gestión
			COALESCE(SUM(bg.costos_directos_ejecutados_usd), 0) as balance_direct_costs_executed,
			COALESCE(SUM(bg.costos_directos_invertidos_usd), 0) as balance_direct_costs_invested,
			COALESCE(SUM(bg.costos_directos_stock_usd), 0) as balance_direct_costs_stock,
			COALESCE(SUM(bg.arriendo_invertidos_usd), 0) as balance_rent_invested,
			COALESCE(SUM(bg.estructura_invertidos_usd), 0) as balance_structure_invested
		FROM dashboard_view dv
		LEFT JOIN (
			SELECT 
				customer_id, project_id, campaign_id, field_id,
				sowed_area, total_hectares
			FROM dashboard_view 
			WHERE row_kind = 'metric'
		) s ON s.customer_id IS NOT DISTINCT FROM dv.customer_id 
			AND s.project_id IS NOT DISTINCT FROM dv.project_id 
			AND s.campaign_id IS NOT DISTINCT FROM dv.campaign_id 
			AND s.field_id IS NOT DISTINCT FROM dv.field_id
		LEFT JOIN (
			SELECT 
				customer_id, project_id, campaign_id, field_id,
				harvested_area, total_hectares
			FROM dashboard_view 
			WHERE row_kind = 'metric'
		) h ON h.customer_id IS NOT DISTINCT FROM dv.customer_id 
			AND h.project_id IS NOT DISTINCT FROM dv.project_id 
			AND h.campaign_id IS NOT DISTINCT FROM dv.campaign_id 
			AND h.field_id IS NOT DISTINCT FROM dv.field_id
		LEFT JOIN (
			SELECT 
				customer_id, project_id, campaign_id,
				executed_costs_usd, budget_total_usd, costs_progress_pct
			FROM dashboard_view 
			WHERE row_kind = 'metric'
		) ca ON ca.customer_id IS NOT DISTINCT FROM dv.customer_id 
			AND ca.project_id IS NOT DISTINCT FROM dv.project_id 
			AND ca.campaign_id IS NOT DISTINCT FROM dv.campaign_id
		LEFT JOIN (
			SELECT 
				customer_id, project_id, campaign_id,
				income_usd, total_invested_usd
			FROM dashboard_view 
			WHERE row_kind = 'metric'
		) o ON o.customer_id IS NOT DISTINCT FROM dv.customer_id 
			AND o.project_id IS NOT DISTINCT FROM dv.project_id 
			AND o.campaign_id IS NOT DISTINCT FROM dv.campaign_id
		LEFT JOIN (
			SELECT 
				customer_id, project_id, campaign_id,
				costos_directos_ejecutados_usd, costos_directos_invertidos_usd, costos_directos_stock_usd,
				arriendo_invertidos_usd, estructura_invertidos_usd
			FROM dashboard_view 
			WHERE row_kind = 'metric'
		) bg ON bg.customer_id IS NOT DISTINCT FROM dv.customer_id 
			AND bg.project_id IS NOT DISTINCT FROM dv.project_id 
			AND bg.campaign_id IS NOT DISTINCT FROM dv.campaign_id
		WHERE dv.row_kind = 'metric'
	`

	// Aplicar filtros si están presentes
	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND dv.customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND dv.project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND dv.campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}
	if len(filter.FieldIDs) > 0 {
		query += " AND dv.field_id = ANY($4)"
		args = append(args, filter.FieldIDs)
	}

	query += " GROUP BY dv.customer_id, dv.project_id, dv.campaign_id, dv.field_id"

	// Ejecutar la consulta
	var result struct {
		SowingHectares             float64 `db:"sowing_hectares"`
		SowingTotalHectares        float64 `db:"sowing_total_hectares"`
		HarvestHectares            float64 `db:"harvest_hectares"`
		HarvestTotalHectares       float64 `db:"harvest_total_hectares"`
		CostsExecutedUSD           float64 `db:"costs_executed_usd"`
		CostsBudgetUSD             float64 `db:"costs_budget_usd"`
		CostsProgressPct           float64 `db:"costs_progress_pct"`
		OperatingIncomeUSD         float64 `db:"operating_income_usd"`
		OperatingTotalCostsUSD     float64 `db:"operating_total_costs_usd"`
		BalanceDirectCostsExecuted float64 `db:"balance_direct_costs_executed"`
		BalanceDirectCostsInvested float64 `db:"balance_direct_costs_invested"`
		BalanceDirectCostsStock    float64 `db:"balance_direct_costs_stock"`
		BalanceRentInvested        float64 `db:"balance_rent_invested"`
		BalanceStructureInvested   float64 `db:"balance_structure_invested"`
	}

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&result).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get dashboard data", err)
	}

	// Obtener datos de cultivos
	crops, err := r.getCropIncidence(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener datos de inversores
	investors, err := r.getInvestorContributions(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Obtener indicadores operativos
	operational, err := r.getOperationalIndicators(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Construir la respuesta
	dashboardData := &domain.DashboardData{
		Metrics: &domain.DashboardMetrics{
			Sowing: &domain.DashboardSowing{
				ProgressPct:   decimal.NewFromFloat(calculateProgress(result.SowingHectares, result.SowingTotalHectares)),
				Hectares:      decimal.NewFromFloat(result.SowingHectares),
				TotalHectares: decimal.NewFromFloat(result.SowingTotalHectares),
			},
			Harvest: &domain.DashboardHarvest{
				ProgressPct:   decimal.NewFromFloat(calculateProgress(result.HarvestHectares, result.HarvestTotalHectares)),
				Hectares:      decimal.NewFromFloat(result.HarvestHectares),
				TotalHectares: decimal.NewFromFloat(result.HarvestTotalHectares),
			},
			Costs: &domain.DashboardCosts{
				ProgressPct: decimal.NewFromFloat(result.CostsProgressPct),
				ExecutedUSD: decimal.NewFromFloat(result.CostsExecutedUSD),
				BudgetUSD:   decimal.NewFromFloat(result.CostsBudgetUSD),
			},
			InvestorContributions: &domain.DashboardInvestorContributions{
				ProgressPct: decimal.NewFromFloat(100), // Siempre 100% por proyecto
				Breakdown:   investors,
			},
			OperatingResult: &domain.DashboardOperatingResult{
				ProgressPct:   decimal.NewFromFloat(calculateOperatingProgress(result.OperatingIncomeUSD, result.OperatingTotalCostsUSD)),
				IncomeUSD:     decimal.NewFromFloat(result.OperatingIncomeUSD),
				TotalCostsUSD: decimal.NewFromFloat(result.OperatingTotalCostsUSD),
			},
		},
		ManagementBalance: &domain.DashboardManagementBalance{
			Summary: &domain.DashboardBalanceSummary{
				IncomeUSD:              decimal.NewFromFloat(result.OperatingIncomeUSD),
				DirectCostsExecutedUSD: decimal.NewFromFloat(result.BalanceDirectCostsExecuted),
				DirectCostsInvestedUSD: decimal.NewFromFloat(result.BalanceDirectCostsInvested),
				StockUSD:               decimal.NewFromFloat(result.BalanceDirectCostsStock),
				RentUSD:                decimal.NewFromFloat(result.BalanceRentInvested),
				StructureUSD:           decimal.NewFromFloat(result.BalanceStructureInvested),
				OperatingResultUSD:     decimal.NewFromFloat(result.OperatingIncomeUSD - result.BalanceDirectCostsExecuted),
				OperatingResultPct:     decimal.NewFromFloat(calculateOperatingProgress(result.OperatingIncomeUSD, result.BalanceDirectCostsExecuted)),
			},
			Breakdown: []domain.DashboardBalanceBreakdown{
				{
					Label:       "Seed",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					StockUSD:    nil,
				},
				{
					Label:       "Supplies",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.Zero,
					StockUSD:    nil,
				},
				{
					Label:       "Labors",
					ExecutedUSD: decimal.NewFromFloat(result.BalanceDirectCostsExecuted),
					InvestedUSD: decimal.Zero,
					StockUSD:    decimalPtr(decimal.NewFromFloat(result.BalanceDirectCostsStock)),
				},
				{
					Label:       "Rent",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.NewFromFloat(result.BalanceRentInvested),
					StockUSD:    decimalPtr(decimal.Zero),
				},
				{
					Label:       "Structure",
					ExecutedUSD: decimal.Zero,
					InvestedUSD: decimal.NewFromFloat(result.BalanceStructureInvested),
					StockUSD:    decimalPtr(decimal.Zero),
				},
			},
			TotalsRow: &domain.DashboardBalanceTotals{
				ExecutedUSD: decimal.NewFromFloat(result.BalanceDirectCostsExecuted),
				InvestedUSD: decimal.NewFromFloat(result.BalanceDirectCostsInvested),
				StockUSD:    decimal.NewFromFloat(result.BalanceDirectCostsStock),
			},
		},
		CropIncidence:         crops,
		OperationalIndicators: operational,
	}

	return dashboardData, nil
}

func (r *Repository) getCropIncidence(ctx context.Context, filter domain.DashboardFilter) (*domain.DashboardCropIncidence, error) {
	query := `
		SELECT 
			crop_name,
			crop_hectares,
			incidence_pct,
			cost_per_ha_usd
		FROM dashboard_view 
		WHERE row_kind = 'metric'
		AND crop_name IS NOT NULL
	`

	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}
	if len(filter.FieldIDs) > 0 {
		query += " AND field_id = ANY($4)"
		args = append(args, filter.FieldIDs)
	}

	query += " ORDER BY crop_name"

	var crops []struct {
		Name         string  `db:"crop_name"`
		Hectares     float64 `db:"crop_hectares"`
		IncidencePct float64 `db:"incidence_pct"`
		CostPerHa    float64 `db:"cost_per_ha_usd"`
	}

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&crops).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get crop incidence data", err)
	}

	// Convertir a dominio
	domainCrops := make([]domain.DashboardCrop, 0, len(crops))
	var totalHectares, totalCosts decimal.Decimal

	for _, crop := range crops {
		hectares := decimal.NewFromFloat(crop.Hectares)
		costPerHa := decimal.NewFromFloat(crop.CostPerHa)
		incidencePct := decimal.NewFromFloat(crop.IncidencePct)

		domainCrops = append(domainCrops, domain.DashboardCrop{
			Name:         crop.Name,
			Hectares:     hectares,
			RotationPct:  incidencePct,
			CostUSDPerHa: costPerHa,
			IncidencePct: incidencePct,
		})

		totalHectares = totalHectares.Add(hectares)
		totalCosts = totalCosts.Add(hectares.Mul(costPerHa))
	}

	// Calcular totales
	var totalRotationPct, totalCostPerHectare decimal.Decimal
	if totalHectares.GreaterThan(decimal.Zero) {
		totalRotationPct = decimal.NewFromFloat(100)
		totalCostPerHectare = totalCosts.Div(totalHectares)
	}

	return &domain.DashboardCropIncidence{
		Crops: domainCrops,
		Total: &domain.DashboardCropTotal{
			Hectares:          totalHectares,
			RotationPct:       totalRotationPct,
			CostUSDPerHectare: totalCostPerHectare,
		},
	}, nil
}

func (r *Repository) getInvestorContributions(ctx context.Context, filter domain.DashboardFilter) ([]domain.DashboardInvestorBreakdown, error) {
	query := `
		SELECT 
			investor_id,
			investor_name,
			investor_percentage_pct
		FROM dashboard_view 
		WHERE row_kind = 'metric'
		AND investor_id IS NOT NULL
	`

	args := []interface{}{}
	if len(filter.CustomerIDs) > 0 {
		query += " AND customer_id = ANY($1)"
		args = append(args, filter.CustomerIDs)
	}
	if len(filter.ProjectIDs) > 0 {
		query += " AND project_id = ANY($2)"
		args = append(args, filter.ProjectIDs)
	}
	if len(filter.CampaignIDs) > 0 {
		query += " AND campaign_id = ANY($3)"
		args = append(args, filter.CampaignIDs)
	}
	if len(filter.FieldIDs) > 0 {
		query += " AND field_id = ANY($4)"
		args = append(args, filter.FieldIDs)
	}

	query += " ORDER BY investor_id"

	var investors []struct {
		InvestorID   int64   `db:"investor_id"`
		InvestorName string  `db:"investor_name"`
		Percentage   float64 `db:"investor_percentage_pct"`
	}

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&investors).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get investor contributions data", err)
	}

	// Convertir a dominio
	domainInvestors := make([]domain.DashboardInvestorBreakdown, 0, len(investors))
	for _, investor := range investors {
		domainInvestors = append(domainInvestors, domain.DashboardInvestorBreakdown{
			InvestorID:   investor.InvestorID,
			InvestorName: investor.InvestorName,
			PercentPct:   decimal.NewFromFloat(investor.Percentage),
		})
	}

	return domainInvestors, nil
}

func (r *Repository) getOperationalIndicators(ctx context.Context, filter domain.DashboardFilter) (*domain.DashboardOperationalIndicators, error) {
	// Crear indicadores operativos básicos
	cards := []domain.DashboardOperationalCard{
		{
			Key:           "first_workorder",
			Title:         "Primera orden de trabajo",
			Date:          nil,
			WorkorderID:   nil,
			WorkorderCode: nil,
			AuditID:       nil,
			AuditCode:     nil,
			Status:        nil,
		},
		{
			Key:           "last_workorder",
			Title:         "Última orden de trabajo",
			Date:          nil,
			WorkorderID:   nil,
			WorkorderCode: nil,
			AuditID:       nil,
			AuditCode:     nil,
			Status:        nil,
		},
		{
			Key:           "last_stock_audit",
			Title:         "Último arqueo de stock",
			Date:          nil,
			WorkorderID:   nil,
			WorkorderCode: nil,
			AuditID:       nil,
			AuditCode:     nil,
			Status:        nil,
		},
		{
			Key:           "campaign_close",
			Title:         "Cierre de campaña",
			Date:          nil,
			WorkorderID:   nil,
			WorkorderCode: nil,
			AuditID:       nil,
			AuditCode:     nil,
			Status:        stringPtr("pending"),
		},
	}

	return &domain.DashboardOperationalIndicators{
		Cards: cards,
	}, nil
}

// Funciones auxiliares para cálculos
func calculateProgress(actual, total float64) float64 {
	if total > 0 {
		return (actual / total) * 100
	}
	return 0
}

func calculateOperatingProgress(income, costs float64) float64 {
	if costs > 0 {
		return ((income - costs) / costs) * 100
	}
	return 0
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func stringPtr(s string) *string {
	return &s
}
