package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
)

// DashboardResponse representa la respuesta del dashboard (DTO viejo)
type DashboardResponse struct {
	Metrics       MetricsResponse       `json:"metrics"`
	Sowing        SowingResponse        `json:"sowing"`
	Harvest       HarvestResponse       `json:"harvest"`
	Costs         CostsResponse         `json:"costs"`
	Contributions ContributionsResponse `json:"contributions"`
}

// MetricsResponse representa las métricas del dashboard
type MetricsResponse struct {
	OperatingResult OperatingResultMetric `json:"operating_result"`
	CropIncidence   []CropIncidence       `json:"crop_incidence"`
}

// OperatingResultMetric representa el resultado operativo
type OperatingResultMetric struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	DirectCostsStockUSD    decimal.Decimal `json:"direct_costs_stock_usd"`
	LeaseInvestedUSD       decimal.Decimal `json:"lease_invested_usd"`
	StructureInvestedUSD   decimal.Decimal `json:"structure_invested_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
	SemillaCost            decimal.Decimal `json:"semilla_cost"`
	InsumosCost            decimal.Decimal `json:"insumos_cost"`
	LaboresCost            decimal.Decimal `json:"labores_cost"`
}

// CropIncidence representa la incidencia de cultivos
type CropIncidence struct {
	CropName     string          `json:"crop_name"`
	CropHectares decimal.Decimal `json:"crop_hectares"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
	CostPerHaUSD decimal.Decimal `json:"cost_per_ha_usd"`
}

// SowingResponse representa la respuesta de siembra
type SowingResponse struct {
	ProgressPercent   decimal.Decimal `json:"progress_percent"`
	TotalHectares     decimal.Decimal `json:"total_hectares"`
	SowedHectares     decimal.Decimal `json:"sowed_hectares"`
	RemainingHectares decimal.Decimal `json:"remaining_hectares"`
}

// HarvestResponse representa la respuesta de cosecha
type HarvestResponse struct {
	ProgressPercent decimal.Decimal `json:"progress_percent"`
	TotalTons       decimal.Decimal `json:"total_tons"`
	HarvestedTons   decimal.Decimal `json:"harvested_tons"`
	RemainingTons   decimal.Decimal `json:"remaining_tons"`
}

// CostsResponse representa la respuesta de costos
type CostsResponse struct {
	ProgressPercent decimal.Decimal `json:"progress_percent"`
	TotalCosts      decimal.Decimal `json:"total_costs"`
	ExecutedCosts   decimal.Decimal `json:"executed_costs"`
	RemainingCosts  decimal.Decimal `json:"remaining_costs"`
}

// ContributionsResponse representa la respuesta de aportes
type ContributionsResponse struct {
	ProgressPercent     decimal.Decimal `json:"progress_percent"`
	TotalPercentage     decimal.Decimal `json:"total_percentage"`
	InvestedPercentage  decimal.Decimal `json:"invested_percentage"`
	RemainingPercentage decimal.Decimal `json:"remaining_percentage"`
}

// FromDashboardData convierte domain.DashboardData a DashboardResponse (DTO viejo)
func FromDashboardData(data *domain.DashboardData) *DashboardResponse {
	metrics := convertMetricsFromDomain(data.Metrics)

	// Llenar el CropIncidence desde data.CropIncidence
	if data.CropIncidence != nil {
		metrics.CropIncidence = convertCropIncidenceFromDomain(data.CropIncidence)
	}

	return &DashboardResponse{
		Metrics:       metrics,
		Sowing:        convertSowingFromDomain(data.Metrics.Sowing),
		Harvest:       convertHarvestFromDomain(data.Metrics.Harvest),
		Costs:         convertCostsFromDomain(data.Metrics.Costs),
		Contributions: convertContributionsFromDomain(data.Metrics.InvestorContributions),
	}
}

// convertMetricsFromDomain convierte las métricas del dominio al DTO
func convertMetricsFromDomain(metrics *domain.DashboardMetrics) MetricsResponse {
	return MetricsResponse{
		OperatingResult: convertOperatingResultFromDomain(metrics.OperatingResult),
		CropIncidence:   []CropIncidence{}, // Inicialmente vacío, se llenará desde el DTO principal
	}
}

// convertOperatingResultFromDomain convierte el resultado operativo del dominio al DTO
func convertOperatingResultFromDomain(op *domain.DashboardOperatingResult) OperatingResultMetric {
	return OperatingResultMetric{
		IncomeUSD:              op.ResultUSD,     // Usar ResultUSD en lugar de IncomeUSD
		DirectCostsExecutedUSD: op.TotalCostsUSD, // Usar TotalCostsUSD
		DirectCostsInvestedUSD: op.TotalCostsUSD, // Usar TotalCostsUSD
		DirectCostsStockUSD:    decimal.Zero,     // No existe en el dominio
		LeaseInvestedUSD:       decimal.Zero,     // No existe en el dominio
		StructureInvestedUSD:   decimal.Zero,     // No existe en el dominio
		OperatingResultUSD:     op.ResultUSD,
		OperatingResultPct:     op.ProgressPct,
		SemillaCost:            decimal.Zero, // No existe en el dominio
		InsumosCost:            decimal.Zero, // No existe en el dominio
		LaboresCost:            decimal.Zero, // No existe en el dominio
	}
}

// convertCropIncidenceFromDomain convierte la incidencia de cultivos del dominio al DTO
func convertCropIncidenceFromDomain(cropIncidence *domain.DashboardCropIncidence) []CropIncidence {
	if cropIncidence == nil || cropIncidence.Crops == nil {
		return []CropIncidence{}
	}

	result := make([]CropIncidence, len(cropIncidence.Crops))
	for i, crop := range cropIncidence.Crops {
		result[i] = CropIncidence{
			CropName:     crop.Name,
			CropHectares: crop.Hectares,
			IncidencePct: crop.IncidencePct,
			CostPerHaUSD: crop.CostUSDPerHa,
		}
	}
	return result
}

// convertSowingFromDomain convierte la siembra del dominio al DTO
func convertSowingFromDomain(sowing *domain.DashboardSowing) SowingResponse {
	if sowing == nil {
		return SowingResponse{}
	}

	remaining := sowing.TotalHectares.Sub(sowing.Hectares)
	if remaining.LessThan(decimal.Zero) {
		remaining = decimal.Zero
	}

	return SowingResponse{
		ProgressPercent:   sowing.ProgressPct,
		TotalHectares:     sowing.TotalHectares,
		SowedHectares:     sowing.Hectares,
		RemainingHectares: remaining,
	}
}

// convertHarvestFromDomain convierte la cosecha del dominio al DTO
func convertHarvestFromDomain(harvest *domain.DashboardHarvest) HarvestResponse {
	if harvest == nil {
		return HarvestResponse{}
	}

	remaining := harvest.TotalHectares.Sub(harvest.Hectares)
	if remaining.LessThan(decimal.Zero) {
		remaining = decimal.Zero
	}

	return HarvestResponse{
		ProgressPercent: harvest.ProgressPct,
		TotalTons:       harvest.TotalHectares, // Usar TotalHectares como aproximación
		HarvestedTons:   harvest.Hectares,      // Usar Hectares como aproximación
		RemainingTons:   remaining,             // Calcular remanente
	}
}

// convertCostsFromDomain convierte los costos del dominio al DTO
func convertCostsFromDomain(costs *domain.DashboardCosts) CostsResponse {
	if costs == nil {
		return CostsResponse{}
	}

	remaining := costs.BudgetUSD.Sub(costs.ExecutedUSD)
	if remaining.LessThan(decimal.Zero) {
		remaining = decimal.Zero
	}

	return CostsResponse{
		ProgressPercent: costs.ProgressPct,
		TotalCosts:      costs.BudgetUSD,
		ExecutedCosts:   costs.ExecutedUSD,
		RemainingCosts:  remaining,
	}
}

// convertContributionsFromDomain convierte los aportes del dominio al DTO
func convertContributionsFromDomain(contributions *domain.DashboardInvestorContributions) ContributionsResponse {
	if contributions == nil {
		return ContributionsResponse{}
	}

	totalPercentage := decimal.Zero
	investedPercentage := decimal.Zero

	// Calcular totales desde el breakdown
	if contributions.Breakdown != nil {
		for _, breakdown := range contributions.Breakdown {
			totalPercentage = totalPercentage.Add(breakdown.PercentPct)
			investedPercentage = investedPercentage.Add(breakdown.PercentPct)
		}
	}

	remainingPercentage := decimal.NewFromInt(100).Sub(investedPercentage)
	if remainingPercentage.LessThan(decimal.Zero) {
		remainingPercentage = decimal.Zero
	}

	return ContributionsResponse{
		ProgressPercent:     contributions.ProgressPct,
		TotalPercentage:     totalPercentage,
		InvestedPercentage:  investedPercentage,
		RemainingPercentage: remainingPercentage,
	}
}
