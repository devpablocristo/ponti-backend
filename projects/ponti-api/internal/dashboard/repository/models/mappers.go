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
		breakdown[i] = domain.DashboardInvestorBreakdown{
			InvestorID:   model.InvestorID,
			InvestorName: model.InvestorName,
			PercentPct:   model.InvestorPercentage,
		}
	}

	return &domain.DashboardInvestorContributions{
		ProgressPct: decimal.NewFromInt(100), // Siempre 100% por proyecto
		Breakdown:   breakdown,
	}
}

// OperatingResultToDomain convierte OperatingResultModel a domain.DashboardOperatingResult
func (m *DashboardModelMapper) OperatingResultToDomain(model *OperatingResultModel) *domain.DashboardOperatingResult {
	if model == nil {
		return &domain.DashboardOperatingResult{}
	}

	return &domain.DashboardOperatingResult{
		ProgressPct:   model.ResultPct,
		IncomeUSD:     model.IncomeUSD,
		TotalCostsUSD: model.TotalCostsUSD,
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
