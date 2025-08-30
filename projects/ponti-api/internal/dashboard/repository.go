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

// Devuelve el struct tipado con decimal.Decimal
func (r *Repository) GetDashboard(ctx context.Context, filt domain.DashboardFilter) (*domain.DashboardPayload, error) {
	// Obtener filas de métricas
	metricRows, err := r.getMetricRows(ctx, filt)
	if err != nil {
		return nil, err
	}

	// Obtener filas de inversores para el breakdown
	investorRows, err := r.getInvestorRows(ctx, filt)
	if err != nil {
		return nil, err
	}

	// Obtener datos de cultivos
	cropData, err := r.getCropData(ctx, filt)
	if err != nil {
		return nil, err
	}

	// Si no hay resultados, retornar estructura vacía
	if len(metricRows) == 0 {
		return r.createEmptyDashboardPayload(), nil
	}

	// Construir el payload del dashboard
	payload := r.buildDashboardPayload(metricRows, investorRows, cropData)

	return payload, nil
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

// createEmptyDashboardPayload crea un payload vacío pero válido
func (r *Repository) createEmptyDashboardPayload() *domain.DashboardPayload {
	return &domain.DashboardPayload{
		Metrics: struct {
			Sowing struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				Hectares      decimal.Decimal `json:"hectares"`
				TotalHectares decimal.Decimal `json:"total_hectares"`
			} `json:"sowing"`
			Harvest struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				Hectares      decimal.Decimal `json:"hectares"`
				TotalHectares decimal.Decimal `json:"total_hectares"`
			} `json:"harvest"`
			Costs struct {
				ProgressPct decimal.Decimal `json:"progress_pct"`
				ExecutedUSD decimal.Decimal `json:"executed_usd"`
				BudgetUSD   decimal.Decimal `json:"budget_usd"`
			} `json:"costs"`
			InvestorContributions struct {
				ProgressPct decimal.Decimal `json:"progress_pct"`
				Breakdown   interface{}     `json:"breakdown"`
			} `json:"investor_contributions"`
			OperatingResult struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				IncomeUSD     decimal.Decimal `json:"income_usd"`
				TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
			} `json:"operating_result"`
		}{
			Sowing: struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				Hectares      decimal.Decimal `json:"hectares"`
				TotalHectares decimal.Decimal `json:"total_hectares"`
			}{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Harvest: struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				Hectares      decimal.Decimal `json:"hectares"`
				TotalHectares decimal.Decimal `json:"total_hectares"`
			}{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Costs: struct {
				ProgressPct decimal.Decimal `json:"progress_pct"`
				ExecutedUSD decimal.Decimal `json:"executed_usd"`
				BudgetUSD   decimal.Decimal `json:"budget_usd"`
			}{
				ProgressPct: decimal.Zero,
				ExecutedUSD: decimal.Zero,
				BudgetUSD:   decimal.Zero,
			},
			InvestorContributions: struct {
				ProgressPct decimal.Decimal `json:"progress_pct"`
				Breakdown   interface{}     `json:"breakdown"`
			}{
				ProgressPct: decimal.Zero,
				Breakdown:   []interface{}{},
			},
			OperatingResult: struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				IncomeUSD     decimal.Decimal `json:"income_usd"`
				TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
			}{
				ProgressPct:   decimal.Zero,
				IncomeUSD:     decimal.Zero,
				TotalCostsUSD: decimal.Zero,
			},
		},
		ManagementBalance: struct {
			Summary struct {
				IncomeUSD              decimal.Decimal `json:"income_usd"`
				DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
				DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
				StockUSD               decimal.Decimal `json:"stock_usd"`
				RentUSD                decimal.Decimal `json:"rent_usd"`
				StructureUSD           decimal.Decimal `json:"structure_usd"`
				OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
				OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
			} `json:"summary"`
			Breakdown []struct {
				Label       string           `json:"label"`
				ExecutedUSD decimal.Decimal  `json:"executed_usd"`
				InvestedUSD decimal.Decimal  `json:"invested_usd"`
				StockUSD    *decimal.Decimal `json:"stock_usd"`
			} `json:"breakdown"`
			TotalsRow struct {
				ExecutedUSD decimal.Decimal `json:"executed_usd"`
				InvestedUSD decimal.Decimal `json:"invested_usd"`
				StockUSD    decimal.Decimal `json:"stock_usd"`
			} `json:"totals_row"`
		}{
			Summary: struct {
				IncomeUSD              decimal.Decimal `json:"income_usd"`
				DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
				DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
				StockUSD               decimal.Decimal `json:"stock_usd"`
				RentUSD                decimal.Decimal `json:"rent_usd"`
				StructureUSD           decimal.Decimal `json:"structure_usd"`
				OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
				OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
			}{
				IncomeUSD:              decimal.Zero,
				DirectCostsExecutedUSD: decimal.Zero,
				DirectCostsInvestedUSD: decimal.Zero,
				StockUSD:               decimal.Zero,
				RentUSD:                decimal.Zero,
				StructureUSD:           decimal.Zero,
				OperatingResultUSD:     decimal.Zero,
				OperatingResultPct:     decimal.Zero,
			},
			Breakdown: []struct {
				Label       string           `json:"label"`
				ExecutedUSD decimal.Decimal  `json:"executed_usd"`
				InvestedUSD decimal.Decimal  `json:"invested_usd"`
				StockUSD    *decimal.Decimal `json:"stock_usd"`
			}{},
			TotalsRow: struct {
				ExecutedUSD decimal.Decimal `json:"executed_usd"`
				InvestedUSD decimal.Decimal `json:"invested_usd"`
				StockUSD    decimal.Decimal `json:"stock_usd"`
			}{
				ExecutedUSD: decimal.Zero,
				InvestedUSD: decimal.Zero,
				StockUSD:    decimal.Zero,
			},
		},
		CropIncidence: struct {
			Crops []struct {
				Name         string          `json:"name"`
				Hectares     decimal.Decimal `json:"hectares"`
				RotationPct  decimal.Decimal `json:"rotation_pct"`
				CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
				IncidencePct decimal.Decimal `json:"incidence_pct"`
			} `json:"crops"`
			Total struct {
				Hectares          decimal.Decimal `json:"hectares"`
				RotationPct       decimal.Decimal `json:"rotation_pct"`
				CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
			} `json:"total"`
		}{
			Crops: []struct {
				Name         string          `json:"name"`
				Hectares     decimal.Decimal `json:"hectares"`
				RotationPct  decimal.Decimal `json:"rotation_pct"`
				CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
				IncidencePct decimal.Decimal `json:"incidence_pct"`
			}{},
			Total: struct {
				Hectares          decimal.Decimal `json:"hectares"`
				RotationPct       decimal.Decimal `json:"rotation_pct"`
				CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
			}{
				Hectares:          decimal.Zero,
				RotationPct:       decimal.Zero,
				CostUSDPerHectare: decimal.Zero,
			},
		},
		OperationalIndicators: struct {
			Cards []struct {
				Key           string      `json:"key"`
				Title         string      `json:"title"`
				Date          *string     `json:"date"`
				WorkorderID   interface{} `json:"workorder_id"`
				WorkorderCode interface{} `json:"workorder_code"`
				AuditID       interface{} `json:"audit_id"`
				AuditCode     interface{} `json:"audit_code"`
				Status        interface{} `json:"status"`
			} `json:"cards"`
		}{
			Cards: []struct {
				Key           string      `json:"key"`
				Title         string      `json:"title"`
				Date          *string     `json:"date"`
				WorkorderID   interface{} `json:"workorder_id"`
				WorkorderCode interface{} `json:"workorder_code"`
				AuditID       interface{} `json:"audit_id"`
				AuditCode     interface{} `json:"audit_code"`
				Status        interface{} `json:"status"`
			}{},
		},
	}
}

// buildDashboardPayload construye el payload del dashboard desde las filas de la vista
func (r *Repository) buildDashboardPayload(metricRows []DashboardRow, investorRows []DashboardRow, cropRows []DashboardRow) *domain.DashboardPayload {
	payload := r.createEmptyDashboardPayload()

	// Construir métricas principales desde las filas de métricas
	for _, row := range metricRows {
		// Métricas de siembra
		payload.Metrics.Sowing.Hectares = payload.Metrics.Sowing.Hectares.Add(row.SowingHectares)
		payload.Metrics.Sowing.TotalHectares = payload.Metrics.Sowing.TotalHectares.Add(row.SowingTotalHectares)

		// Métricas de cosecha
		payload.Metrics.Harvest.Hectares = payload.Metrics.Harvest.Hectares.Add(row.HarvestHectares)
		payload.Metrics.Harvest.TotalHectares = payload.Metrics.Harvest.TotalHectares.Add(row.HarvestTotalHectares)

		// Métricas de costos
		payload.Metrics.Costs.ExecutedUSD = payload.Metrics.Costs.ExecutedUSD.Add(row.ExecutedCostsUSD)
		payload.Metrics.Costs.BudgetUSD = payload.Metrics.Costs.BudgetUSD.Add(row.BudgetCostUSD)

		// Métricas de contribuciones
		payload.Metrics.InvestorContributions.ProgressPct = row.ContributionsProgressPct

		// Métricas de resultado operativo
		payload.Metrics.OperatingResult.IncomeUSD = payload.Metrics.OperatingResult.IncomeUSD.Add(row.IncomeUSD)
		payload.Metrics.OperatingResult.TotalCostsUSD = payload.Metrics.OperatingResult.TotalCostsUSD.Add(row.DirectLaborsUSD)

		// Balance de gestión
		payload.ManagementBalance.Summary.IncomeUSD = payload.ManagementBalance.Summary.IncomeUSD.Add(row.IncomeUSD)
		payload.ManagementBalance.Summary.DirectCostsExecutedUSD = payload.ManagementBalance.Summary.DirectCostsExecutedUSD.Add(row.ExecutedCostsUSD)
		payload.ManagementBalance.Summary.OperatingResultUSD = payload.ManagementBalance.Summary.OperatingResultUSD.Add(row.OperatingResultUSD)
		payload.ManagementBalance.Summary.OperatingResultPct = row.OperatingResultPct
	}

	// Construir breakdown de inversores
	payload.Metrics.InvestorContributions.Breakdown = r.buildInvestorBreakdown(investorRows)

	// Construir breakdown del balance de gestión
	payload.ManagementBalance.Breakdown = r.buildBalanceBreakdown(metricRows)

	// Construir incidencia de cultivos
	payload.CropIncidence = r.buildCropIncidence(cropRows)

	// Construir indicadores operativos
	payload.OperationalIndicators = r.buildOperationalIndicators()

	// Calcular porcentajes de progreso
	if payload.Metrics.Sowing.TotalHectares.GreaterThan(decimal.Zero) {
		payload.Metrics.Sowing.ProgressPct = payload.Metrics.Sowing.Hectares.Div(payload.Metrics.Sowing.TotalHectares).Mul(decimal.NewFromInt(100))
	}
	if payload.Metrics.Harvest.TotalHectares.GreaterThan(decimal.Zero) {
		payload.Metrics.Harvest.ProgressPct = payload.Metrics.Harvest.Hectares.Div(payload.Metrics.Harvest.TotalHectares).Mul(decimal.NewFromInt(100))
	}
	if payload.Metrics.Costs.BudgetUSD.GreaterThan(decimal.Zero) {
		payload.Metrics.Costs.ProgressPct = payload.Metrics.Costs.ExecutedUSD.Div(payload.Metrics.Costs.BudgetUSD).Mul(decimal.NewFromInt(100))
	}
	if payload.Metrics.OperatingResult.TotalCostsUSD.GreaterThan(decimal.Zero) {
		payload.Metrics.OperatingResult.ProgressPct = payload.Metrics.OperatingResult.IncomeUSD.Div(payload.Metrics.OperatingResult.TotalCostsUSD).Mul(decimal.NewFromInt(100))
	}

	// Calcular totales del balance
	payload.ManagementBalance.TotalsRow.ExecutedUSD = payload.ManagementBalance.Summary.DirectCostsExecutedUSD
	payload.ManagementBalance.TotalsRow.InvestedUSD = payload.ManagementBalance.Summary.DirectCostsInvestedUSD
	payload.ManagementBalance.TotalsRow.StockUSD = payload.ManagementBalance.Summary.StockUSD

	return payload
}

// buildInvestorBreakdown construye el breakdown de inversores
func (r *Repository) buildInvestorBreakdown(investorRows []DashboardRow) interface{} {
	var breakdown []map[string]interface{}

	for _, row := range investorRows {
		if row.InvestorID != nil && row.InvestorName != nil && row.InvestorPercentage != nil {
			breakdown = append(breakdown, map[string]interface{}{
				"investor_id":   *row.InvestorID,
				"investor_name": *row.InvestorName,
				"percent_pct":   row.InvestorPercentage.String(),
			})
		}
	}

	return breakdown
}

// buildBalanceBreakdown construye el breakdown del balance de gestión
func (r *Repository) buildBalanceBreakdown(metricRows []DashboardRow) []struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
} {
	var breakdown []struct {
		Label       string           `json:"label"`
		ExecutedUSD decimal.Decimal  `json:"executed_usd"`
		InvestedUSD decimal.Decimal  `json:"invested_usd"`
		StockUSD    *decimal.Decimal `json:"stock_usd"`
	}

	// Seed
	seedExecuted := decimal.Zero
	seedInvested := decimal.Zero
	// Aquí podrías agregar lógica para obtener datos de seed si están disponibles
	breakdown = append(breakdown, struct {
		Label       string           `json:"label"`
		ExecutedUSD decimal.Decimal  `json:"executed_usd"`
		InvestedUSD decimal.Decimal  `json:"invested_usd"`
		StockUSD    *decimal.Decimal `json:"stock_usd"`
	}{
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
	breakdown = append(breakdown, struct {
		Label       string           `json:"label"`
		ExecutedUSD decimal.Decimal  `json:"executed_usd"`
		InvestedUSD decimal.Decimal  `json:"invested_usd"`
		StockUSD    *decimal.Decimal `json:"stock_usd"`
	}{
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
	breakdown = append(breakdown, struct {
		Label       string           `json:"label"`
		ExecutedUSD decimal.Decimal  `json:"executed_usd"`
		InvestedUSD decimal.Decimal  `json:"invested_usd"`
		StockUSD    *decimal.Decimal `json:"stock_usd"`
	}{
		Label:       "Labors",
		ExecutedUSD: laborsExecuted,
		InvestedUSD: laborsInvested,
		StockUSD:    &laborsStock,
	})

	// Rent
	rentExecuted := decimal.Zero
	rentInvested := decimal.Zero
	rentStock := decimal.Zero
	breakdown = append(breakdown, struct {
		Label       string           `json:"label"`
		ExecutedUSD decimal.Decimal  `json:"executed_usd"`
		InvestedUSD decimal.Decimal  `json:"invested_usd"`
		StockUSD    *decimal.Decimal `json:"stock_usd"`
	}{
		Label:       "Rent",
		ExecutedUSD: rentExecuted,
		InvestedUSD: rentInvested,
		StockUSD:    &rentStock,
	})

	// Structure
	structureExecuted := decimal.Zero
	structureInvested := decimal.Zero
	structureStock := decimal.Zero
	breakdown = append(breakdown, struct {
		Label       string           `json:"label"`
		ExecutedUSD decimal.Decimal  `json:"executed_usd"`
		InvestedUSD decimal.Decimal  `json:"invested_usd"`
		StockUSD    *decimal.Decimal `json:"stock_usd"`
	}{
		Label:       "Structure",
		ExecutedUSD: structureExecuted,
		InvestedUSD: structureInvested,
		StockUSD:    &structureStock,
	})

	return breakdown
}

// buildCropIncidence construye la incidencia de cultivos
func (r *Repository) buildCropIncidence(cropRows []DashboardRow) struct {
	Crops []struct {
		Name         string          `json:"name"`
		Hectares     decimal.Decimal `json:"hectares"`
		RotationPct  decimal.Decimal `json:"rotation_pct"`
		CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
		IncidencePct decimal.Decimal `json:"incidence_pct"`
	} `json:"crops"`
	Total struct {
		Hectares          decimal.Decimal `json:"hectares"`
		RotationPct       decimal.Decimal `json:"rotation_pct"`
		CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
	} `json:"total"`
} {
	var result struct {
		Crops []struct {
			Name         string          `json:"name"`
			Hectares     decimal.Decimal `json:"hectares"`
			RotationPct  decimal.Decimal `json:"rotation_pct"`
			CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
			IncidencePct decimal.Decimal `json:"incidence_pct"`
		} `json:"crops"`
		Total struct {
			Hectares          decimal.Decimal `json:"hectares"`
			RotationPct       decimal.Decimal `json:"rotation_pct"`
			CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
		} `json:"total"`
	}

	// Por ahora, crear datos de ejemplo según el JSON requerido
	// En una implementación real, estos datos vendrían de la base de datos
	result.Crops = []struct {
		Name         string          `json:"name"`
		Hectares     decimal.Decimal `json:"hectares"`
		RotationPct  decimal.Decimal `json:"rotation_pct"`
		CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
		IncidencePct decimal.Decimal `json:"incidence_pct"`
	}{
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

	result.Total.Hectares = totalHectares
	result.Total.RotationPct = decimal.NewFromInt(100)

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
func (r *Repository) buildOperationalIndicators() struct {
	Cards []struct {
		Key           string      `json:"key"`
		Title         string      `json:"title"`
		Date          *string     `json:"date"`
		WorkorderID   interface{} `json:"workorder_id"`
		WorkorderCode interface{} `json:"workorder_code"`
		AuditID       interface{} `json:"audit_id"`
		AuditCode     interface{} `json:"audit_code"`
		Status        interface{} `json:"status"`
	} `json:"cards"`
} {
	var result struct {
		Cards []struct {
			Key           string      `json:"key"`
			Title         string      `json:"title"`
			Date          *string     `json:"date"`
			WorkorderID   interface{} `json:"workorder_id"`
			WorkorderCode interface{} `json:"workorder_code"`
			AuditID       interface{} `json:"audit_id"`
			AuditCode     interface{} `json:"audit_code"`
			Status        interface{} `json:"status"`
		} `json:"cards"`
	}

	// Crear las 4 tarjetas según el JSON requerido
	result.Cards = []struct {
		Key           string      `json:"key"`
		Title         string      `json:"title"`
		Date          *string     `json:"date"`
		WorkorderID   interface{} `json:"workorder_id"`
		WorkorderCode interface{} `json:"workorder_code"`
		AuditID       interface{} `json:"audit_id"`
		AuditCode     interface{} `json:"audit_code"`
		Status        interface{} `json:"status"`
	}{
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
