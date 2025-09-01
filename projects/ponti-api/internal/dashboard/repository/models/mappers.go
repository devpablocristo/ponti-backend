package models

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	"github.com/shopspring/decimal"
)

// DashboardModelMapper mapea entre modelos y dominio
type DashboardModelMapper struct {
	config *DashboardConfig
}

// NewDashboardModelMapper crea una nueva instancia del mapper
func NewDashboardModelMapper() *DashboardModelMapper {
	return &DashboardModelMapper{
		config: GetDefaultDashboardConfig(),
	}
}

// NewDashboardModelMapperWithConfig crea un mapper con configuración personalizada
func NewDashboardModelMapperWithConfig(config *DashboardConfig) *DashboardModelMapper {
	return &DashboardModelMapper{
		config: config,
	}
}

// ToDomain convierte DashboardModel a dominio
func (m *DashboardModelMapper) ToDomain(model *DashboardModel) *domain.Dashboard {
	if model == nil {
		return nil
	}

	return &domain.Dashboard{
		ID: model.ID,
		Base: shareddomain.Base{
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
			CreatedBy: model.CreatedBy,
			UpdatedBy: model.UpdatedBy,
		},
	}
}

// FromDomain convierte dominio a DashboardModel
func (m *DashboardModelMapper) FromDomain(domain *domain.Dashboard) *DashboardModel {
	if domain == nil {
		return nil
	}

	return &DashboardModel{
		ID: domain.ID,
		Base: sharedmodels.Base{
			CreatedBy: domain.CreatedBy,
			UpdatedBy: domain.UpdatedBy,
		},
	}
}

// DashboardDataToDomain convierte DashboardDataModel a dominio
func (m *DashboardModelMapper) DashboardDataToDomain(
	data *DashboardDataModel,
	crops []CropIncidenceModel,
	investors []InvestorContributionModel,
	operational *OperationalIndicatorModel,
) *domain.DashboardData {
	if data == nil {
		return nil
	}

	// Convertir cultivos
	domainCrops := make([]domain.DashboardCrop, 0, len(crops))
	var totalHectares, totalCosts decimal.Decimal

	for _, crop := range crops {
		domainCrops = append(domainCrops, domain.DashboardCrop{
			Name:         crop.Name,
			Hectares:     crop.Hectares,
			RotationPct:  crop.IncidencePct,
			CostUSDPerHa: crop.CostPerHa,
			IncidencePct: crop.IncidencePct,
		})

		totalHectares = totalHectares.Add(crop.Hectares)
		totalCosts = totalCosts.Add(crop.Hectares.Mul(crop.CostPerHa))
	}

	// Calcular totales de cultivos
	var totalRotationPct, totalCostPerHectare decimal.Decimal
	if totalHectares.GreaterThan(decimal.Zero) {
		totalRotationPct = decimal.NewFromFloat(100)
		totalCostPerHectare = totalCosts.Div(totalHectares)
	}

	// Convertir inversores
	domainInvestors := make([]domain.DashboardInvestorBreakdown, 0, len(investors))
	for _, investor := range investors {
		domainInvestors = append(domainInvestors, domain.DashboardInvestorBreakdown{
			InvestorID:   investor.InvestorID,
			InvestorName: investor.InvestorName,
			PercentPct:   investor.Percentage,
		})
	}

	// Convertir indicadores operativos usando configuración
	var operationalCards []domain.DashboardOperationalCard
	if operational != nil {
		operationalCards = []domain.DashboardOperationalCard{
			{
				Key:           m.config.CardKeys.FirstWorkorder,
				Title:         m.config.CardTitles.FirstWorkorder,
				Date:          operational.PrimeraOrdenFecha,
				WorkorderID:   operational.PrimeraOrdenID,
				WorkorderCode: nil,
				AuditID:       nil,
				AuditCode:     nil,
				Status:        nil,
			},
			{
				Key:           m.config.CardKeys.LastWorkorder,
				Title:         m.config.CardTitles.LastWorkorder,
				Date:          operational.UltimaOrdenFecha,
				WorkorderID:   operational.UltimaOrdenID,
				WorkorderCode: nil,
				AuditID:       nil,
				AuditCode:     nil,
				Status:        nil,
			},
			{
				Key:           m.config.CardKeys.LastStockAudit,
				Title:         m.config.CardTitles.LastStockAudit,
				Date:          operational.ArqueoStockFecha,
				WorkorderID:   nil,
				WorkorderCode: nil,
				AuditID:       nil,
				AuditCode:     nil,
				Status:        nil,
			},
			{
				Key:           m.config.CardKeys.CampaignClose,
				Title:         m.config.CardTitles.CampaignClose,
				Date:          operational.CierreCampanaFecha,
				WorkorderID:   nil,
				WorkorderCode: nil,
				AuditID:       nil,
				AuditCode:     nil,
				Status:        stringPtr(m.config.DefaultStatuses.CampaignClose),
			},
		}
	}

	// Calcular progreso de inversores basado en datos reales
	investorProgressPct := calculateInvestorProgress(investors)

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
				ProgressPct: investorProgressPct,
				Breakdown:   domainInvestors,
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
			Breakdown: []domain.DashboardBalanceBreakdown{
				{
					Label:       m.config.BalanceLabels.Seed,
					ExecutedUSD: data.SemillaEjecutadosUSD,
					InvestedUSD: data.SemillaInvertidosUSD,
					StockUSD:    decimalPtr(data.SemillaStockUSD),
				},
				{
					Label:       m.config.BalanceLabels.Supplies,
					ExecutedUSD: data.InsumosEjecutadosUSD,
					InvestedUSD: data.InsumosInvertidosUSD,
					StockUSD:    decimalPtr(data.InsumosStockUSD),
				},
				{
					Label:       m.config.BalanceLabels.Labors,
					ExecutedUSD: data.LaboresEjecutadosUSD,
					InvestedUSD: data.LaboresInvertidosUSD,
					StockUSD:    decimalPtr(data.LaboresStockUSD),
				},
				{
					Label:       m.config.BalanceLabels.Rent,
					ExecutedUSD: decimal.Zero, // No se calcula según la vista
					InvestedUSD: data.ArriendoInvertidosUSD,
					StockUSD:    nil, // No se calcula según la vista
				},
				{
					Label:       m.config.BalanceLabels.Structure,
					ExecutedUSD: decimal.Zero, // No se calcula según la vista
					InvestedUSD: data.EstructuraInvertidosUSD,
					StockUSD:    nil, // No se calcula según la vista
				},
			},
			TotalsRow: &domain.DashboardBalanceTotals{
				ExecutedUSD: data.CostosDirectosEjecutados,
				InvestedUSD: data.CostosDirectosInvertidos,
				StockUSD:    data.CostosDirectosStock,
			},
		},
		CropIncidence: &domain.DashboardCropIncidence{
			Crops: domainCrops,
			Total: &domain.DashboardCropTotal{
				Hectares:          totalHectares,
				RotationPct:       totalRotationPct,
				CostUSDPerHectare: totalCostPerHectare,
			},
		},
		OperationalIndicators: &domain.DashboardOperationalIndicators{
			Cards: operationalCards,
		},
	}
}

// calculateInvestorProgress calcula el progreso real de inversores basado en datos
func calculateInvestorProgress(investors []InvestorContributionModel) decimal.Decimal {
	if len(investors) == 0 {
		return decimal.Zero
	}

	var totalPercentage decimal.Decimal
	for _, investor := range investors {
		totalPercentage = totalPercentage.Add(investor.Percentage)
	}

	// Si la suma es 100%, retornar 100, sino retornar el valor real
	if totalPercentage.Equal(decimal.NewFromFloat(100)) {
		return decimal.NewFromFloat(100)
	}
	return totalPercentage
}

// Funciones auxiliares
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func stringPtr(s string) *string {
	return &s
}
