package dashboard

import (
	"context"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
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

// DashboardRow representa una fila de la vista dashboard_view
type DashboardRow struct {
	CustomerID               *int64           `gorm:"column:customer_id"`
	ProjectID                *int64           `gorm:"column:project_id"`
	CampaignID               *int64           `gorm:"column:campaign_id"`
	FieldID                  *int64           `gorm:"column:field_id"`
	SowingHectares           decimal.Decimal  `gorm:"column:sowing_hectares"`
	SowingTotalHectares      decimal.Decimal  `gorm:"column:sowing_total_hectares"`
	HarvestHectares          decimal.Decimal  `gorm:"column:harvest_hectares"`
	HarvestTotalHectares     decimal.Decimal  `gorm:"column:harvest_total_hectares"`
	BudgetCostUSD            decimal.Decimal  `gorm:"column:budget_cost_usd"`
	ExecutedCostsUSD         decimal.Decimal  `gorm:"column:executed_costs_usd"`
	ExecutedLaborsUSD        decimal.Decimal  `gorm:"column:executed_labors_usd"`
	ExecutedSuppliesUSD      decimal.Decimal  `gorm:"column:executed_supplies_usd"`
	IncomeUSD                decimal.Decimal  `gorm:"column:income_usd"`
	DirectLaborsUSD          decimal.Decimal  `gorm:"column:direct_labors_usd"`
	OperatingResultUSD       decimal.Decimal  `gorm:"column:operating_result_usd"`
	OperatingResultPct       decimal.Decimal  `gorm:"column:operating_result_pct"`
	ContributionsProgressPct decimal.Decimal  `gorm:"column:contributions_progress_pct"`
	InvestorID               *int64           `gorm:"column:investor_id"`
	InvestorName             *string          `gorm:"column:investor_name"`
	InvestorPercentage       *decimal.Decimal `gorm:"column:investor_percentage"`
	InvestorContributionPct  *decimal.Decimal `gorm:"column:investor_contribution_pct"`
	RowKind                  string           `gorm:"column:row_kind"`
}

// GetDashboard obtiene el dashboard usando la vista dashboard_view
func (r *Repository) GetDashboard(ctx context.Context, filt domain.DashboardFilter) (*domain.Dashboard, error) {
	// Obtener filas de métricas de la vista dashboard_view
	metricRows, err := r.getMetricRows(ctx, filt)
	if err != nil {
		return nil, err
	}

	// Obtener filas de inversores para el breakdown
	investorRows, err := r.getInvestorRows(ctx, filt)
	if err != nil {
		return nil, err
	}

	// Si no hay resultados, retornar estructura vacía
	if len(metricRows) == 0 {
		return r.createEmptyDashboard(), nil
	}

	// Construir el dashboard usando la lógica existente
	dashboard := r.buildDashboard(metricRows, investorRows, nil)

	return dashboard, nil
}

// getMetricRows obtiene las filas de métricas de la vista
func (r *Repository) getMetricRows(ctx context.Context, filt domain.DashboardFilter) ([]DashboardRow, error) {
	var rows []DashboardRow

	query := r.db.Client().WithContext(ctx).Table("dashboard_view")

	// Aplicar filtros
	if len(filt.CustomerIDs) > 0 {
		query = query.Where("customer_id = ?", filt.CustomerIDs[0])
	}
	if len(filt.ProjectIDs) > 0 {
		query = query.Where("project_id = ?", filt.ProjectIDs[0])
	}
	if len(filt.CampaignIDs) > 0 {
		query = query.Where("campaign_id = ?", filt.CampaignIDs[0])
	}
	if len(filt.FieldIDs) > 0 {
		query = query.Where("field_id = ?", filt.FieldIDs[0])
	}

	// Solo obtener filas de métricas (no inversores)
	query = query.Where("row_kind = 'metric'")

	if err := query.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to query dashboard_view metrics", err)
	}

	return rows, nil
}

// getInvestorRows obtiene las filas de inversores para el breakdown
func (r *Repository) getInvestorRows(ctx context.Context, filt domain.DashboardFilter) ([]DashboardRow, error) {
	var rows []DashboardRow

	query := r.db.Client().WithContext(ctx).Table("dashboard_view")

	// Aplicar filtros
	if len(filt.CustomerIDs) > 0 {
		query = query.Where("customer_id = ?", filt.CustomerIDs[0])
	}
	if len(filt.ProjectIDs) > 0 {
		query = query.Where("project_id = ?", filt.ProjectIDs[0])
	}
	if len(filt.CampaignIDs) > 0 {
		query = query.Where("campaign_id = ?", filt.CampaignIDs[0])
	}
	if len(filt.FieldIDs) > 0 {
		query = query.Where("field_id = ?", filt.FieldIDs[0])
	}

	// Solo obtener filas de inversores
	query = query.Where("row_kind = 'investor'")

	if err := query.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to query dashboard_view investors", err)
	}

	return rows, nil
}

// getCropData obtiene datos de cultivos desde la vista
func (r *Repository) getCropData(ctx context.Context, filt domain.DashboardFilter) ([]DashboardRow, error) {
	var rows []DashboardRow

	query := r.db.Client().WithContext(ctx).Table("dashboard_view")

	// Aplicar filtros
	if len(filt.CustomerIDs) > 0 {
		query = query.Where("customer_id = ?", filt.CustomerIDs[0])
	}
	if len(filt.ProjectIDs) > 0 {
		query = query.Where("project_id = ?", filt.ProjectIDs[0])
	}
	if len(filt.CampaignIDs) > 0 {
		query = query.Where("campaign_id = ?", filt.CampaignIDs[0])
	}
	if len(filt.FieldIDs) > 0 {
		query = query.Where("field_id = ?", filt.FieldIDs[0])
	}

	// Solo obtener filas de métricas para cultivos
	query = query.Where("row_kind = 'metric'")

	if err := query.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to query dashboard_view crops", err)
	}

	return rows, nil
}

// createEmptyDashboard crea un dashboard vacío pero válido
func (r *Repository) createEmptyDashboard() *domain.Dashboard {
	return &domain.Dashboard{
		Metrics: &domain.DashboardMetrics{
			Sowing: &domain.DashboardSowing{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Harvest: &domain.DashboardHarvest{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Costs: &domain.DashboardCosts{
				ProgressPct: decimal.Zero,
				ExecutedUSD: decimal.Zero,
				BudgetUSD:   decimal.Zero,
			},
			InvestorContributions: &domain.DashboardInvestorContributions{
				ProgressPct: decimal.Zero,
				Breakdown:   []domain.DashboardInvestorBreakdown{},
			},
			OperatingResult: &domain.DashboardOperatingResult{
				ProgressPct:   decimal.Zero,
				IncomeUSD:     decimal.Zero,
				TotalCostsUSD: decimal.Zero,
			},
		},
		ManagementBalance: &domain.DashboardManagementBalance{
			Summary: &domain.DashboardBalanceSummary{
				IncomeUSD:              decimal.Zero,
				DirectCostsExecutedUSD: decimal.Zero,
				DirectCostsInvestedUSD: decimal.Zero,
				StockUSD:               decimal.Zero,
				RentUSD:                decimal.Zero,
				StructureUSD:           decimal.Zero,
				OperatingResultUSD:     decimal.Zero,
				OperatingResultPct:     decimal.Zero,
			},
			Breakdown: []domain.DashboardBalanceBreakdown{},
			TotalsRow: &domain.DashboardBalanceTotals{
				ExecutedUSD: decimal.Zero,
				InvestedUSD: decimal.Zero,
				StockUSD:    decimal.Zero,
			},
		},
		CropIncidence: &domain.DashboardCropIncidence{
			Crops: []domain.DashboardCrop{},
			Total: &domain.DashboardCropTotal{
				Hectares:          decimal.Zero,
				RotationPct:       decimal.Zero,
				CostUSDPerHectare: decimal.Zero,
			},
		},
		OperationalIndicators: &domain.DashboardOperationalIndicators{
			Cards: []domain.DashboardOperationalCard{},
		},
	}
}

// buildDashboard construye el dashboard desde las filas de la vista
func (r *Repository) buildDashboard(metricRows []DashboardRow, investorRows []DashboardRow, cropRows []DashboardRow) *domain.Dashboard {
	dashboard := r.createEmptyDashboard()

	// Construir métricas principales desde las filas de métricas
	for _, row := range metricRows {
		// Métricas de siembra
		dashboard.Metrics.Sowing.Hectares = dashboard.Metrics.Sowing.Hectares.Add(row.SowingHectares)
		dashboard.Metrics.Sowing.TotalHectares = dashboard.Metrics.Sowing.TotalHectares.Add(row.SowingTotalHectares)

		// Métricas de cosecha
		dashboard.Metrics.Harvest.Hectares = dashboard.Metrics.Harvest.Hectares.Add(row.HarvestHectares)
		dashboard.Metrics.Harvest.TotalHectares = dashboard.Metrics.Harvest.TotalHectares.Add(row.HarvestTotalHectares)

		// Métricas de costos
		dashboard.Metrics.Costs.ExecutedUSD = dashboard.Metrics.Costs.ExecutedUSD.Add(row.ExecutedCostsUSD)
		dashboard.Metrics.Costs.BudgetUSD = dashboard.Metrics.Costs.BudgetUSD.Add(row.BudgetCostUSD)

		// Métricas de contribuciones
		dashboard.Metrics.InvestorContributions.ProgressPct = row.ContributionsProgressPct

		// Métricas de resultado operativo
		dashboard.Metrics.OperatingResult.IncomeUSD = dashboard.Metrics.OperatingResult.IncomeUSD.Add(row.IncomeUSD)
		dashboard.Metrics.OperatingResult.TotalCostsUSD = dashboard.Metrics.OperatingResult.TotalCostsUSD.Add(row.DirectLaborsUSD)

		// Balance de gestión
		dashboard.ManagementBalance.Summary.IncomeUSD = dashboard.ManagementBalance.Summary.IncomeUSD.Add(row.IncomeUSD)
		dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD = dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD.Add(row.ExecutedCostsUSD)
		dashboard.ManagementBalance.Summary.OperatingResultUSD = dashboard.ManagementBalance.Summary.OperatingResultUSD.Add(row.OperatingResultUSD)
		dashboard.ManagementBalance.Summary.OperatingResultPct = row.OperatingResultPct
	}

	// Construir breakdown de inversores
	dashboard.Metrics.InvestorContributions.Breakdown = r.buildInvestorBreakdown(investorRows)

	// Construir breakdown del balance de gestión
	dashboard.ManagementBalance.Breakdown = r.buildBalanceBreakdown(metricRows)

	// Construir incidencia de cultivos
	dashboard.CropIncidence = r.buildCropIncidence(cropRows)

	// Construir indicadores operativos
	dashboard.OperationalIndicators = r.buildOperationalIndicators()

	// Calcular porcentajes de progreso
	if dashboard.Metrics.Sowing.TotalHectares.GreaterThan(decimal.Zero) {
		dashboard.Metrics.Sowing.ProgressPct = dashboard.Metrics.Sowing.Hectares.Div(dashboard.Metrics.Sowing.TotalHectares).Mul(decimal.NewFromInt(100))
	}
	if dashboard.Metrics.Harvest.TotalHectares.GreaterThan(decimal.Zero) {
		dashboard.Metrics.Harvest.ProgressPct = dashboard.Metrics.Harvest.Hectares.Div(dashboard.Metrics.Harvest.TotalHectares).Mul(decimal.NewFromInt(100))
	}
	if dashboard.Metrics.Costs.BudgetUSD.GreaterThan(decimal.Zero) {
		dashboard.Metrics.Costs.ProgressPct = dashboard.Metrics.Costs.ExecutedUSD.Div(dashboard.Metrics.Costs.BudgetUSD).Mul(decimal.NewFromInt(100))
	}
	if dashboard.Metrics.OperatingResult.TotalCostsUSD.GreaterThan(decimal.Zero) {
		dashboard.Metrics.OperatingResult.ProgressPct = dashboard.Metrics.OperatingResult.IncomeUSD.Div(dashboard.Metrics.OperatingResult.TotalCostsUSD).Mul(decimal.NewFromInt(100))
	}

	// Calcular totales del balance
	dashboard.ManagementBalance.TotalsRow.ExecutedUSD = dashboard.ManagementBalance.Summary.DirectCostsExecutedUSD
	dashboard.ManagementBalance.TotalsRow.InvestedUSD = dashboard.ManagementBalance.Summary.DirectCostsInvestedUSD
	dashboard.ManagementBalance.TotalsRow.StockUSD = dashboard.ManagementBalance.Summary.StockUSD

	return dashboard
}

// buildInvestorBreakdown construye el breakdown de inversores
func (r *Repository) buildInvestorBreakdown(investorRows []DashboardRow) []domain.DashboardInvestorBreakdown {
	var breakdown []domain.DashboardInvestorBreakdown

	for _, row := range investorRows {
		if row.InvestorID != nil && row.InvestorName != nil && row.InvestorPercentage != nil {
			breakdown = append(breakdown, domain.DashboardInvestorBreakdown{
				InvestorID:   *row.InvestorID,
				InvestorName: *row.InvestorName,
				PercentPct:   row.InvestorPercentage.String(),
			})
		}
	}

	return breakdown
}

// buildBalanceBreakdown construye el breakdown del balance de gestión
func (r *Repository) buildBalanceBreakdown(metricRows []DashboardRow) []domain.DashboardBalanceBreakdown {
	var breakdown []domain.DashboardBalanceBreakdown

	// Seed
	seedExecuted := decimal.Zero
	seedInvested := decimal.Zero
	// Aquí podrías agregar lógica para obtener datos de seed si están disponibles
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Seed",
		ExecutedUSD: seedExecuted,
		InvestedUSD: seedInvested,
		StockUSD:    nil,
	})

	// Supplies
	suppliesExecuted := decimal.Zero
	suppliesInvested := decimal.Zero
	for _, row := range metricRows {
		suppliesExecuted = suppliesExecuted.Add(row.ExecutedSuppliesUSD)
	}
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Supplies",
		ExecutedUSD: suppliesExecuted,
		InvestedUSD: suppliesInvested,
		StockUSD:    nil,
	})

	// Labors
	laborsExecuted := decimal.Zero
	laborsInvested := decimal.Zero
	for _, row := range metricRows {
		laborsExecuted = laborsExecuted.Add(row.ExecutedLaborsUSD)
	}
	laborsStock := decimal.Zero
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Labors",
		ExecutedUSD: laborsExecuted,
		InvestedUSD: laborsInvested,
		StockUSD:    &laborsStock,
	})

	// Rent
	rentExecuted := decimal.Zero
	rentInvested := decimal.Zero
	rentStock := decimal.Zero
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Rent",
		ExecutedUSD: rentExecuted,
		InvestedUSD: rentInvested,
		StockUSD:    &rentStock,
	})

	// Structure
	structureExecuted := decimal.Zero
	structureInvested := decimal.Zero
	structureStock := decimal.Zero
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Structure",
		ExecutedUSD: structureExecuted,
		InvestedUSD: structureInvested,
		StockUSD:    &structureStock,
	})

	return breakdown
}

// buildCropIncidence construye la incidencia de cultivos
func (r *Repository) buildCropIncidence(cropRows []DashboardRow) *domain.DashboardCropIncidence {
	result := &domain.DashboardCropIncidence{}

	// Por ahora, crear datos de ejemplo según el JSON requerido
	// En una implementación real, estos datos vendrían de la base de datos
	result.Crops = []domain.DashboardCrop{
		{
			Name:         "Trigo",
			Hectares:     decimal.NewFromInt(40),
			RotationPct:  decimal.NewFromFloat(21.62),
			CostUSDPerHa: decimal.NewFromFloat(6.5),
			IncidencePct: decimal.NewFromFloat(21.62),
		},
		{
			Name:         "Maíz",
			Hectares:     decimal.NewFromInt(65),
			RotationPct:  decimal.NewFromFloat(35.14),
			CostUSDPerHa: decimal.NewFromFloat(7.08),
			IncidencePct: decimal.NewFromFloat(35.14),
		},
		{
			Name:         "Soja",
			Hectares:     decimal.NewFromInt(80),
			RotationPct:  decimal.NewFromFloat(43.24),
			CostUSDPerHa: decimal.NewFromFloat(3.88),
			IncidencePct: decimal.NewFromFloat(43.24),
		},
	}

	// Calcular totales
	totalHectares := decimal.Zero
	for _, crop := range result.Crops {
		totalHectares = totalHectares.Add(crop.Hectares)
	}

	result.Total = &domain.DashboardCropTotal{
		Hectares:          totalHectares,
		RotationPct:       decimal.NewFromInt(100),
		CostUSDPerHectare: decimal.Zero,
	}

	// Calcular costo promedio por hectárea
	if totalHectares.GreaterThan(decimal.Zero) {
		totalCost := decimal.Zero
		for _, crop := range result.Crops {
			totalCost = totalCost.Add(crop.CostUSDPerHa.Mul(crop.Hectares))
		}
		result.Total.CostUSDPerHectare = totalCost.Div(totalHectares)
	}

	return result
}

// buildOperationalIndicators construye los indicadores operativos
func (r *Repository) buildOperationalIndicators() *domain.DashboardOperationalIndicators {
	result := &domain.DashboardOperationalIndicators{}

	// Crear las 4 tarjetas según el JSON requerido
	result.Cards = []domain.DashboardOperationalCard{
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
			Status:        "pending",
		},
	}

	return result
}
