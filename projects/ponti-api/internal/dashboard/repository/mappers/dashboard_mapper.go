package mappers

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== MAPPER FUNCTIONS =====

// ToDashboard convierte modelos de base de datos a entidad de dominio Dashboard
func ToDashboard(cards []models.DashboardCards, balance []models.DashboardBalance, cropIncidence []models.DashboardCropIncidence, operationalIndicators []models.DashboardOperationalIndicators) *domain.Dashboard {
	return &domain.Dashboard{
		Metrics:               buildMetrics(cards),
		ManagementBalance:     buildManagementBalance(balance),
		CropIncidence:         buildCropIncidence(cropIncidence),
		OperationalIndicators: buildOperationalIndicators(operationalIndicators),
	}
}

// buildMetrics construye las métricas del dashboard
func buildMetrics(cards []models.DashboardCards) *domain.DashboardMetrics {
	if len(cards) == 0 {
		return createEmptyMetrics()
	}

	card := cards[0] // Tomamos la primera tarjeta como base

	return &domain.DashboardMetrics{
		Sowing: &domain.DashboardSowing{
			ProgressPct:   card.SowingProgressPct,
			Hectares:      card.SowedArea,
			TotalHectares: card.TotalHectares,
		},
		Harvest: &domain.DashboardHarvest{
			ProgressPct:   card.HarvestProgressPct,
			Hectares:      card.HarvestedArea,
			TotalHectares: card.TotalHectares,
		},
		Costs: &domain.DashboardCosts{
			ProgressPct: card.CostsProgressPct,
			ExecutedUSD: card.LaborsExecutedUSD.Add(card.SuppliesExecutedUSD).Add(card.SeedExecutedUSD),
			BudgetUSD:   card.BudgetCostUSD,
		},
		InvestorContributions: &domain.DashboardInvestorContributions{
			ProgressPct: card.ContributionsProgressPct,
			Breakdown:   []domain.DashboardInvestorBreakdown{},
		},
		OperatingResult: &domain.DashboardOperatingResult{
			ProgressPct:   card.OperatingResultPct,
			IncomeUSD:     card.IncomeUSD,
			TotalCostsUSD: card.OperatingResultUSD,
		},
	}
}

// buildManagementBalance construye el balance de gestión
func buildManagementBalance(balance []models.DashboardBalance) *domain.DashboardManagementBalance {
	if len(balance) == 0 {
		return createEmptyManagementBalance()
	}

	bal := balance[0] // Tomamos el primer balance como base

	return &domain.DashboardManagementBalance{
		Summary: &domain.DashboardBalanceSummary{
			IncomeUSD:              decimal.Zero, // No disponible en el modelo actual
			DirectCostsExecutedUSD: bal.DirectCostsExecutedUSD,
			DirectCostsInvestedUSD: bal.DirectCostsInvestedUSD,
			StockUSD:               bal.StockUSD,
			RentUSD:                bal.RentUSD,
			StructureUSD:           bal.StructureUSD,
			OperatingResultUSD:     decimal.Zero, // Calculado
			OperatingResultPct:     decimal.Zero, // Calculado
		},
		Breakdown: buildBalanceBreakdown(bal),
		TotalsRow: &domain.DashboardBalanceTotals{
			ExecutedUSD: bal.DirectCostsExecutedUSD,
			InvestedUSD: bal.DirectCostsInvestedUSD,
			StockUSD:    bal.StockUSD,
		},
	}
}

// buildBalanceBreakdown construye el breakdown del balance
func buildBalanceBreakdown(bal models.DashboardBalance) []domain.DashboardBalanceBreakdown {
	var breakdown []domain.DashboardBalanceBreakdown

	// Seed
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Seed",
		ExecutedUSD: bal.SeedExecutedUSD,
		InvestedUSD: bal.SeedInvestedUSD,
		StockUSD:    nil,
	})

	// Supplies
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Supplies",
		ExecutedUSD: bal.SuppliesExecutedUSD,
		InvestedUSD: bal.SuppliesInvestedUSD,
		StockUSD:    nil,
	})

	// Labors
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Labors",
		ExecutedUSD: bal.LaborsExecutedUSD,
		InvestedUSD: bal.LaborsInvestedUSD,
		StockUSD:    &bal.StockUSD,
	})

	// Rent
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Rent",
		ExecutedUSD: decimal.Zero,
		InvestedUSD: decimal.Zero,
		StockUSD:    &bal.RentUSD,
	})

	// Structure
	breakdown = append(breakdown, domain.DashboardBalanceBreakdown{
		Label:       "Structure",
		ExecutedUSD: decimal.Zero,
		InvestedUSD: decimal.Zero,
		StockUSD:    &bal.StructureUSD,
	})

	return breakdown
}

// buildCropIncidence construye la incidencia de cultivos
func buildCropIncidence(cropIncidence []models.DashboardCropIncidence) *domain.DashboardCropIncidence {
	if len(cropIncidence) == 0 {
		return createEmptyCropIncidence()
	}

	var crops []domain.DashboardCrop
	var totalHectares, totalRotationPct, totalCostUSD decimal.Decimal

	for _, crop := range cropIncidence {
		crops = append(crops, domain.DashboardCrop{
			Name:         crop.CropName,
			Hectares:     crop.SurfaceHas,
			RotationPct:  crop.RotationPct,
			CostUSDPerHa: crop.CostUSDPerHa,
			IncidencePct: crop.IncidencePct,
		})

		totalHectares = totalHectares.Add(crop.SurfaceHas)
		totalRotationPct = totalRotationPct.Add(crop.RotationPct)
		totalCostUSD = totalCostUSD.Add(crop.TotalCostUSD)
	}

	// Calcular costos por hectárea total
	var costUSDPerHectare decimal.Decimal
	if !totalHectares.IsZero() {
		costUSDPerHectare = totalCostUSD.Div(totalHectares)
	}

	return &domain.DashboardCropIncidence{
		Crops: crops,
		Total: &domain.DashboardCropTotal{
			Hectares:          totalHectares,
			RotationPct:       totalRotationPct,
			CostUSDPerHectare: costUSDPerHectare,
		},
	}
}

// buildOperationalIndicators construye los indicadores operativos
func buildOperationalIndicators(indicators []models.DashboardOperationalIndicators) *domain.DashboardOperationalIndicators {
	if len(indicators) == 0 {
		return createEmptyOperationalIndicators()
	}

	ind := indicators[0] // Tomamos el primer indicador como base

	var cards []domain.DashboardOperationalCard

	// Primera orden de trabajo
	if ind.FirstWorkorderDate != nil {
		dateStr := ind.FirstWorkorderDate.Format("2006-01-02")
		cards = append(cards, domain.DashboardOperationalCard{
			Key:           "first_workorder",
			Title:         "Primera orden de trabajo",
			Date:          &dateStr,
			WorkorderID:   ind.FirstWorkorderID,
			WorkorderCode: nil,
			AuditID:       nil,
			AuditCode:     nil,
			Status:        nil,
		})
	}

	// Última orden de trabajo
	if ind.LastWorkorderDate != nil {
		dateStr := ind.LastWorkorderDate.Format("2006-01-02")
		cards = append(cards, domain.DashboardOperationalCard{
			Key:           "last_workorder",
			Title:         "Última orden de trabajo",
			Date:          &dateStr,
			WorkorderID:   ind.LastWorkorderID,
			WorkorderCode: nil,
			AuditID:       nil,
			AuditCode:     nil,
			Status:        nil,
		})
	}

	// Último arqueo de stock
	if ind.LastStockAuditDate != nil {
		dateStr := ind.LastStockAuditDate.Format("2006-01-02")
		cards = append(cards, domain.DashboardOperationalCard{
			Key:           "last_stock_audit",
			Title:         "Último arqueo de stock",
			Date:          &dateStr,
			WorkorderID:   nil,
			WorkorderCode: nil,
			AuditID:       nil, // No disponible en el modelo actual
			AuditCode:     nil, // No disponible en el modelo actual
			Status:        nil,
		})
	}

	// Cierre de campaña
	cards = append(cards, domain.DashboardOperationalCard{
		Key:           "campaign_close",
		Title:         "Cierre de campaña",
		Date:          nil,
		WorkorderID:   nil,
		WorkorderCode: nil,
		AuditID:       nil,
		AuditCode:     nil,
		Status:        "pending",
	})

	return &domain.DashboardOperationalIndicators{
		Cards: cards,
	}
}

// ===== HELPER FUNCTIONS =====

// createEmptyMetrics crea métricas vacías
func createEmptyMetrics() *domain.DashboardMetrics {
	return &domain.DashboardMetrics{
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
	}
}

// createEmptyManagementBalance crea un balance de gestión vacío
func createEmptyManagementBalance() *domain.DashboardManagementBalance {
	return &domain.DashboardManagementBalance{
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
	}
}

// createEmptyCropIncidence crea una incidencia de cultivos vacía
func createEmptyCropIncidence() *domain.DashboardCropIncidence {
	return &domain.DashboardCropIncidence{
		Crops: []domain.DashboardCrop{},
		Total: &domain.DashboardCropTotal{
			Hectares:          decimal.Zero,
			RotationPct:       decimal.Zero,
			CostUSDPerHectare: decimal.Zero,
		},
	}
}

// createEmptyOperationalIndicators crea indicadores operativos vacíos
func createEmptyOperationalIndicators() *domain.DashboardOperationalIndicators {
	return &domain.DashboardOperationalIndicators{
		Cards: []domain.DashboardOperationalCard{},
	}
}
