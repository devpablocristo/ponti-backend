package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
)

// DashboardModelMapper maneja la conversión entre modelos de base de datos y dominio
type DashboardModelMapper struct{}

// NewDashboardModelMapper crea una nueva instancia del mapper
func NewDashboardModelMapper() *DashboardModelMapper {
	return &DashboardModelMapper{}
}

// DashboardDataToDomain convierte DashboardDataModel a domain.DashboardData
func (m *DashboardModelMapper) DashboardDataToDomain(
	data *DashboardDataModel,
	crops []CropIncidenceModel,
	investors []InvestorContributionModel,
	operational *OperationalIndicatorModel,
) *domain.DashboardData {
	if data == nil {
		return &domain.DashboardData{}
	}

	return &domain.DashboardData{
		Metrics: &domain.DashboardMetrics{
			Sowing: &domain.DashboardSowing{
				ProgressPct:   data.SowingProgressPercent,
				Hectares:      data.SowingHectares,
				TotalHectares: data.SowingTotalHectares,
			},
			Harvest: &domain.DashboardHarvest{
				ProgressPct:   data.HarvestProgressPercent,
				Hectares:      data.HarvestHectares,
				TotalHectares: data.HarvestTotalHectares,
			},
			Costs: &domain.DashboardCosts{
				ProgressPct: data.CostsProgressPct,
				ExecutedUSD: data.CostsExecutedUSD,
				BudgetUSD:   data.CostsBudgetUSD,
			},
			InvestorContributions: &domain.DashboardInvestorContributions{
				ProgressPct: decimal.NewFromInt(100), // Siempre 100% por proyecto
				Breakdown:   m.investorContributionsToDomain(investors),
			},
			OperatingResult: &domain.DashboardOperatingResult{
				ProgressPct:   data.OperatingResultPct,
				IncomeUSD:     data.OperatingIncomeUSD,
				TotalCostsUSD: data.OperatingTotalCostsUSD,
			},
		},
		ManagementBalance: &domain.DashboardManagementBalance{
			Summary: &domain.DashboardBalanceSummary{
				IncomeUSD:              data.OperatingIncomeUSD,
				DirectCostsExecutedUSD: data.CostosDirectosEjecutados,
				DirectCostsInvestedUSD: data.CostosDirectosInvertidos,
				StockUSD:               data.CostosDirectosStock,
				RentUSD:                data.ArriendoInvertidosUSD,
				StructureUSD:           data.EstructuraInvertidosUSD,
				OperatingResultUSD:     data.OperatingResultUSD,
				OperatingResultPct:     data.OperatingResultPct,
			},
			Breakdown: m.managementBalanceBreakdownFromData(data),
			TotalsRow: &domain.DashboardBalanceTotals{
				ExecutedUSD: data.CostosDirectosEjecutados,
				InvestedUSD: data.CostosDirectosInvertidos,
				StockUSD:    data.CostosDirectosStock,
			},
		},
		CropIncidence: &domain.DashboardCropIncidence{
			Crops: m.cropIncidenceToDomain(crops),
			Total: nil, // TODO: Implementar cuando se requiera
		},
		OperationalIndicators: m.operationalIndicatorsToDomain(operational),
	}
}

// SowingProgressToDomain convierte SowingProgressModel a domain.DashboardSowing
func (m *DashboardModelMapper) SowingProgressToDomain(model *SowingProgressModel) *domain.DashboardSowing {
	if model == nil {
		return &domain.DashboardSowing{}
	}

	// Manejar punteros a decimal.Decimal
	var hectares, totalHectares, progressPct decimal.Decimal

	if model.Hectares != nil {
		hectares = *model.Hectares
	} else {
		hectares = decimal.Zero
	}

	if model.TotalHectares != nil {
		totalHectares = *model.TotalHectares
	} else {
		totalHectares = decimal.Zero
	}

	if model.ProgressPct != nil {
		progressPct = *model.ProgressPct
	} else {
		progressPct = decimal.Zero
	}

	return &domain.DashboardSowing{
		ProgressPct:   progressPct,
		Hectares:      hectares,
		TotalHectares: totalHectares,
	}
}

// HarvestProgressToDomain convierte HarvestProgressModel a domain.DashboardHarvest
func (m *DashboardModelMapper) HarvestProgressToDomain(model *HarvestProgressModel) *domain.DashboardHarvest {
	if model == nil {
		return &domain.DashboardHarvest{}
	}

	// Manejar punteros a decimal.Decimal
	var hectares, totalHectares, progressPct decimal.Decimal

	if model.Hectares != nil {
		hectares = *model.Hectares
	} else {
		hectares = decimal.Zero
	}

	if model.TotalHectares != nil {
		totalHectares = *model.TotalHectares
	} else {
		totalHectares = decimal.Zero
	}

	if model.ProgressPct != nil {
		progressPct = *model.ProgressPct
	} else {
		progressPct = decimal.Zero
	}

	return &domain.DashboardHarvest{
		ProgressPct:   progressPct,
		Hectares:      hectares,
		TotalHectares: totalHectares,
	}
}

// CostsProgressToDomain convierte CostsProgressModel a domain.DashboardCosts
func (m *DashboardModelMapper) CostsProgressToDomain(model *CostsProgressModel) *domain.DashboardCosts {
	if model == nil {
		return &domain.DashboardCosts{}
	}

	// Manejar punteros a decimal.Decimal
	var progressPct, executedUSD, budgetUSD decimal.Decimal

	if model.ProgressPct != nil {
		progressPct = *model.ProgressPct
	} else {
		progressPct = decimal.Zero
	}

	if model.ExecutedCostsUSD != nil {
		executedUSD = *model.ExecutedCostsUSD
	} else {
		executedUSD = decimal.Zero
	}

	if model.BudgetTotalUSD != nil {
		budgetUSD = *model.BudgetTotalUSD
	} else {
		budgetUSD = decimal.Zero
	}

	return &domain.DashboardCosts{
		ProgressPct: progressPct,
		ExecutedUSD: executedUSD,
		BudgetUSD:   budgetUSD,
	}
}

// ContributionsProgressToDomain convierte ContributionsProgressModel a domain.DashboardInvestorContributions
func (m *DashboardModelMapper) ContributionsProgressToDomain(models []ContributionsProgressModel) *domain.DashboardInvestorContributions {
	if len(models) == 0 {
		return &domain.DashboardInvestorContributions{
			ProgressPct: decimal.NewFromInt(100),
			Breakdown:   []domain.DashboardInvestorBreakdown{},
		}
	}

	breakdown := make([]domain.DashboardInvestorBreakdown, len(models))
	for i, model := range models {
		// Manejar punteros
		var investorID int64
		var investorName string
		var percentage decimal.Decimal

		if model.InvestorID != nil {
			investorID = *model.InvestorID
		}

		if model.InvestorName != nil {
			investorName = *model.InvestorName
		}

		if model.InvestorPercentage != nil {
			percentage = *model.InvestorPercentage
		} else {
			percentage = decimal.Zero
		}

		breakdown[i] = domain.DashboardInvestorBreakdown{
			InvestorID:   investorID,
			InvestorName: investorName,
			PercentPct:   percentage,
		}
	}

	// Calcular el progreso total (promedio de todos los inversores)
	var totalProgress decimal.Decimal
	if len(models) > 0 {
		sum := decimal.Zero
		count := decimal.Zero
		for _, model := range models {
			if model.ContributionsProgressPct != nil {
				sum = sum.Add(*model.ContributionsProgressPct)
				count = count.Add(decimal.NewFromInt(1))
			}
		}
		if count.GreaterThan(decimal.Zero) {
			totalProgress = sum.Div(count)
		} else {
			totalProgress = decimal.Zero
		}
	} else {
		totalProgress = decimal.Zero
	}

	return &domain.DashboardInvestorContributions{
		ProgressPct: totalProgress,
		Breakdown:   breakdown,
	}
}

// ContributionsProgressToInvestorContribution convierte ContributionsProgressModel a InvestorContributionModel
func (m *DashboardModelMapper) ContributionsProgressToInvestorContribution(models []ContributionsProgressModel) []InvestorContributionModel {
	if len(models) == 0 {
		return []InvestorContributionModel{}
	}

	result := make([]InvestorContributionModel, len(models))
	for i, model := range models {
		// Manejar punteros
		var investorID int64
		var investorName string
		var percentage decimal.Decimal

		if model.InvestorID != nil {
			investorID = *model.InvestorID
		}

		if model.InvestorName != nil {
			investorName = *model.InvestorName
		}

		if model.InvestorPercentage != nil {
			percentage = *model.InvestorPercentage
		} else {
			percentage = decimal.Zero
		}

		result[i] = InvestorContributionModel{
			InvestorID:   investorID,
			InvestorName: investorName,
			Percentage:   percentage,
		}
	}

	return result
}

// OperatingResultToDomain convierte OperatingResultModel a domain.DashboardOperatingResult
func (m *DashboardModelMapper) OperatingResultToDomain(model *OperatingResultModel) *domain.DashboardOperatingResult {
	if model == nil {
		return &domain.DashboardOperatingResult{}
	}

	// Manejar punteros a decimal.Decimal
	var progressPct, incomeUSD, totalCostsUSD decimal.Decimal

	if model.ResultPct != nil {
		progressPct = *model.ResultPct
	} else {
		progressPct = decimal.Zero
	}

	if model.IncomeUSD != nil {
		incomeUSD = *model.IncomeUSD
	} else {
		incomeUSD = decimal.Zero
	}

	if model.TotalCostsUSD != nil {
		totalCostsUSD = *model.TotalCostsUSD
	} else {
		totalCostsUSD = decimal.Zero
	}

	return &domain.DashboardOperatingResult{
		ProgressPct:   progressPct,
		IncomeUSD:     incomeUSD,
		TotalCostsUSD: totalCostsUSD,
	}
}

// ManagementBalanceToDomain convierte ManagementBalanceModel a domain.DashboardManagementBalance
func (m *DashboardModelMapper) ManagementBalanceToDomain(model *ManagementBalanceModel) *domain.DashboardManagementBalance {
	if model == nil {
		return &domain.DashboardManagementBalance{}
	}

	return &domain.DashboardManagementBalance{
		Summary:   m.managementBalanceSummaryToDomain(model.Summary),
		Breakdown: m.managementBalanceBreakdownToDomain(model.Breakdown),
		TotalsRow: m.managementBalanceTotalsToDomain(model.TotalsRow),
	}
}

// cropIncidenceToDomain convierte CropIncidenceModel a domain.DashboardCropBreakdown
func (m *DashboardModelMapper) cropIncidenceToDomain(models []CropIncidenceModel) []domain.DashboardCropBreakdown {
	if len(models) == 0 {
		return []domain.DashboardCropBreakdown{}
	}

	breakdown := make([]domain.DashboardCropBreakdown, len(models))
	for i, model := range models {
		breakdown[i] = domain.DashboardCropBreakdown{
			Name:         model.Name,
			Hectares:     model.Hectares,
			RotationPct:  decimal.Zero, // TODO: Implementar cuando se requiera
			CostUSDPerHa: model.CostPerHa,
			IncidencePct: model.IncidencePct,
		}
	}

	return breakdown
}

// investorContributionsToDomain convierte InvestorContributionModel a domain.DashboardInvestorBreakdown
func (m *DashboardModelMapper) investorContributionsToDomain(models []InvestorContributionModel) []domain.DashboardInvestorBreakdown {
	if len(models) == 0 {
		return []domain.DashboardInvestorBreakdown{}
	}

	breakdown := make([]domain.DashboardInvestorBreakdown, len(models))
	for i, model := range models {
		breakdown[i] = domain.DashboardInvestorBreakdown{
			InvestorID:   model.InvestorID,
			InvestorName: model.InvestorName,
			PercentPct:   model.Percentage,
		}
	}

	return breakdown
}

// operationalIndicatorsToDomain convierte OperationalIndicatorModel a domain.DashboardOperationalIndicators
func (m *DashboardModelMapper) operationalIndicatorsToDomain(model *OperationalIndicatorModel) *domain.DashboardOperationalIndicators {
	if model == nil {
		return &domain.DashboardOperationalIndicators{
			Cards: []domain.DashboardOperationalCard{},
		}
	}

	// TODO: Implementar conversión cuando se requiera
	return &domain.DashboardOperationalIndicators{
		Cards: []domain.DashboardOperationalCard{},
	}
}

// managementBalanceSummaryToDomain convierte ManagementBalanceSummary a domain.DashboardBalanceSummary
func (m *DashboardModelMapper) managementBalanceSummaryToDomain(model *ManagementBalanceSummary) *domain.DashboardBalanceSummary {
	if model == nil {
		return &domain.DashboardBalanceSummary{}
	}

	return &domain.DashboardBalanceSummary{
		IncomeUSD:              model.IncomeUSD,
		DirectCostsExecutedUSD: model.DirectCostsExecutedUSD,
		DirectCostsInvestedUSD: model.DirectCostsInvestedUSD,
		StockUSD:               model.StockUSD,
		RentUSD:                model.RentUSD,
		StructureUSD:           model.StructureUSD,
		OperatingResultUSD:     model.OperatingResultUSD,
		OperatingResultPct:     model.OperatingResultPct,
	}
}

// managementBalanceBreakdownToDomain convierte ManagementBalanceBreakdown a domain.DashboardBalanceBreakdown
func (m *DashboardModelMapper) managementBalanceBreakdownToDomain(models []ManagementBalanceBreakdown) []domain.DashboardBalanceBreakdown {
	if len(models) == 0 {
		return []domain.DashboardBalanceBreakdown{}
	}

	breakdown := make([]domain.DashboardBalanceBreakdown, len(models))
	for i, model := range models {
		breakdown[i] = domain.DashboardBalanceBreakdown{
			Label:       model.Category,
			ExecutedUSD: model.ExecutedUSD,
			InvestedUSD: model.InvestedUSD,
			StockUSD:    &model.StockUSD,
		}
	}

	return breakdown
}

// managementBalanceTotalsToDomain convierte ManagementBalanceTotals a domain.DashboardBalanceTotals
func (m *DashboardModelMapper) managementBalanceTotalsToDomain(model *ManagementBalanceTotals) *domain.DashboardBalanceTotals {
	if model == nil {
		return &domain.DashboardBalanceTotals{}
	}

	return &domain.DashboardBalanceTotals{
		ExecutedUSD: model.TotalExecutedUSD,
		InvestedUSD: model.TotalInvestedUSD,
		StockUSD:    model.TotalStockUSD,
	}
}

// managementBalanceBreakdownFromData convierte DashboardDataModel a domain.DashboardBalanceBreakdown
func (m *DashboardModelMapper) managementBalanceBreakdownFromData(data *DashboardDataModel) []domain.DashboardBalanceBreakdown {
	if data == nil {
		return []domain.DashboardBalanceBreakdown{}
	}

	zeroStock := decimal.Zero

	return []domain.DashboardBalanceBreakdown{
		{
			Label:       "Semilla",
			ExecutedUSD: data.SemillaEjecutadosUSD,
			InvestedUSD: data.SemillaInvertidosUSD,
			StockUSD:    &data.SemillaStockUSD,
		},
		{
			Label:       "Insumos",
			ExecutedUSD: data.InsumosEjecutadosUSD,
			InvestedUSD: data.InsumosInvertidosUSD,
			StockUSD:    &data.InsumosStockUSD,
		},
		{
			Label:       "Labores",
			ExecutedUSD: data.LaboresEjecutadosUSD,
			InvestedUSD: data.LaboresInvertidosUSD,
			StockUSD:    &data.LaboresStockUSD,
		},
		{
			Label:       "Arriendo",
			ExecutedUSD: decimal.Zero,
			InvestedUSD: data.ArriendoInvertidosUSD,
			StockUSD:    &zeroStock,
		},
		{
			Label:       "Estructura",
			ExecutedUSD: decimal.Zero,
			InvestedUSD: data.EstructuraInvertidosUSD,
			StockUSD:    &zeroStock,
		},
	}
}
