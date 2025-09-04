package dto

import (
	"encoding/json"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== REQUEST DTOs =====

// DashboardFilterRequest representa el filtro de request para el dashboard
type DashboardFilterRequest struct {
	CustomerIDs []int64 `json:"customer_ids" binding:"omitempty"`
	ProjectIDs  []int64 `json:"project_ids" binding:"omitempty"`
	CampaignIDs []int64 `json:"campaign_ids" binding:"omitempty"`
	FieldIDs    []int64 `json:"field_ids" binding:"omitempty"`
}

// DashboardRequest representa un request de dashboard (para casos de creación/actualización)
type DashboardRequest struct {
	Metrics               MetricsRequest               `json:"metrics"`
	ManagementBalance     ManagementBalanceRequest     `json:"management_balance"`
	CropIncidence         CropIncidenceRequest         `json:"crop_incidence"`
	OperationalIndicators OperationalIndicatorsRequest `json:"operational_indicators"`
}

// MetricsRequest representa las métricas en el request
type MetricsRequest struct {
	Sowing                SowingMetricRequest          `json:"sowing"`
	Harvest               HarvestMetricRequest         `json:"harvest"`
	Costs                 CostsMetricRequest           `json:"costs"`
	InvestorContributions InvestorContributionsRequest `json:"investor_contributions"`
	OperatingResult       OperatingResultMetricRequest `json:"operating_result"`
}

// SowingMetricRequest representa la métrica de siembra en el request
type SowingMetricRequest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// HarvestMetricRequest representa la métrica de cosecha en el request
type HarvestMetricRequest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// CostsMetricRequest representa la métrica de costos en el request
type CostsMetricRequest struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	BudgetUSD   decimal.Decimal `json:"budget_usd"`
}

// InvestorContributionsRequest representa la métrica de contribuciones en el request
type InvestorContributionsRequest struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	Breakdown   interface{}     `json:"breakdown"`
}

// OperatingResultMetricRequest representa la métrica de resultado operativo en el request
type OperatingResultMetricRequest struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	IncomeUSD     decimal.Decimal `json:"income_usd"`
	TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
}

// ManagementBalanceRequest representa el balance de gestión en el request
type ManagementBalanceRequest struct {
	Summary   BalanceSummaryRequest     `json:"summary"`
	Breakdown []BalanceBreakdownRequest `json:"breakdown"`
	TotalsRow BalanceTotalsRequest      `json:"totals_row"`
}

// BalanceSummaryRequest representa el resumen del balance en el request
type BalanceSummaryRequest struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	StockUSD               decimal.Decimal `json:"stock_usd"`
	RentUSD                decimal.Decimal `json:"rent_usd"`
	StructureUSD           decimal.Decimal `json:"structure_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
}

// BalanceBreakdownRequest representa el desglose por categoría en el request
type BalanceBreakdownRequest struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
}

// BalanceTotalsRequest representa la fila de totales en el request
type BalanceTotalsRequest struct {
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	InvestedUSD decimal.Decimal `json:"invested_usd"`
	StockUSD    decimal.Decimal `json:"stock_usd"`
}

// CropIncidenceRequest representa la incidencia de cultivos en el request
type CropIncidenceRequest struct {
	Crops []CropRequest    `json:"crops"`
	Total CropTotalRequest `json:"total"`
}

// CropRequest representa un cultivo en el request
type CropRequest struct {
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	RotationPct  decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// CropTotalRequest representa los totales de cultivos en el request
type CropTotalRequest struct {
	Hectares          decimal.Decimal `json:"hectares"`
	RotationPct       decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
}

// OperationalIndicatorsRequest representa los indicadores operativos en el request
type OperationalIndicatorsRequest struct {
	Cards []OperationalCardRequest `json:"cards"`
}

// OperationalCardRequest representa una tarjeta de indicador operativo en el request
type OperationalCardRequest struct {
	Key           string      `json:"key"`
	Title         string      `json:"title"`
	Date          *string     `json:"date"`
	WorkorderID   interface{} `json:"workorder_id"`
	WorkorderCode interface{} `json:"workorder_code"`
	AuditID       interface{} `json:"audit_id"`
	AuditCode     interface{} `json:"audit_code"`
	Status        interface{} `json:"status"`
}

// ===== RESPONSE DTOs =====

// roundTo3Decimals redondea un decimal a 3 decimales de precisión
func roundTo3Decimals(d decimal.Decimal) decimal.Decimal {
	return d.Round(3)
}

// DashboardResponse representa la respuesta del dashboard con la estructura exacta del JSON
type DashboardResponse struct {
	Metrics               Metrics               `json:"metrics"`
	ManagementBalance     ManagementBalance     `json:"management_balance"`
	CropIncidence         CropIncidence         `json:"crop_incidence"`
	OperationalIndicators OperationalIndicators `json:"operational_indicators"`
}

// Metrics representa las 5 cards de métricas
type Metrics struct {
	Sowing                SowingMetric          `json:"sowing"`
	Harvest               HarvestMetric         `json:"harvest"`
	Costs                 CostsMetric           `json:"costs"`
	InvestorContributions InvestorContributions `json:"investor_contributions"`
	OperatingResult       OperatingResultMetric `json:"operating_result"`
}

// SowingMetric representa la métrica de siembra
type SowingMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// HarvestMetric representa la métrica de cosecha
type HarvestMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	Hectares      decimal.Decimal `json:"hectares"`
	TotalHectares decimal.Decimal `json:"total_hectares"`
}

// CostsMetric representa la métrica de costos
type CostsMetric struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	BudgetUSD   decimal.Decimal `json:"budget_usd"`
}

// InvestorContributions representa la métrica de contribuciones de inversores
type InvestorContributions struct {
	ProgressPct decimal.Decimal `json:"progress_pct"`
	Breakdown   interface{}     `json:"breakdown"`
}

// OperatingResultMetric representa la métrica de resultado operativo
type OperatingResultMetric struct {
	ProgressPct   decimal.Decimal `json:"progress_pct"`
	IncomeUSD     decimal.Decimal `json:"income_usd"`
	TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
}

// ManagementBalance representa el balance de gestión
type ManagementBalance struct {
	Summary   BalanceSummary     `json:"summary"`
	Breakdown []BalanceBreakdown `json:"breakdown"`
	TotalsRow BalanceTotals      `json:"totals_row"`
}

// BalanceSummary representa el resumen del balance
type BalanceSummary struct {
	IncomeUSD              decimal.Decimal `json:"income_usd"`
	DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
	DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
	StockUSD               decimal.Decimal `json:"stock_usd"`
	RentUSD                decimal.Decimal `json:"rent_usd"`
	StructureUSD           decimal.Decimal `json:"structure_usd"`
	OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
	OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
}

// BalanceBreakdown representa el desglose por categoría
type BalanceBreakdown struct {
	Label       string           `json:"label"`
	ExecutedUSD decimal.Decimal  `json:"executed_usd"`
	InvestedUSD decimal.Decimal  `json:"invested_usd"`
	StockUSD    *decimal.Decimal `json:"stock_usd"`
}

// BalanceTotals representa la fila de totales
type BalanceTotals struct {
	ExecutedUSD decimal.Decimal `json:"executed_usd"`
	InvestedUSD decimal.Decimal `json:"invested_usd"`
	StockUSD    decimal.Decimal `json:"stock_usd"`
}

// CropIncidence representa la incidencia por cultivo
type CropIncidence struct {
	Crops []CropData `json:"crops"`
	Total CropTotal  `json:"total"`
}

// CropData representa los datos de un cultivo específico
type CropData struct {
	Name         string          `json:"name"`
	Hectares     decimal.Decimal `json:"hectares"`
	RotationPct  decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
	IncidencePct decimal.Decimal `json:"incidence_pct"`
}

// CropTotal representa los totales de cultivos
type CropTotal struct {
	Hectares          decimal.Decimal `json:"hectares"`
	RotationPct       decimal.Decimal `json:"rotation_pct"`
	CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
}

// OperationalIndicators representa los indicadores operativos
type OperationalIndicators struct {
	Cards []OperationalCard `json:"cards"`
}

// OperationalCard representa una tarjeta de indicador operativo
type OperationalCard struct {
	Key           string     `json:"key"`
	Title         string     `json:"title"`
	Date          *time.Time `json:"date"`
	WorkorderID   *int64     `json:"workorder_id"`
	WorkorderCode *string    `json:"workorder_code"`
	AuditID       *int64     `json:"audit_id"`
	AuditCode     *string    `json:"audit_code"`
	Status        *string    `json:"status"`
}

// FromDashboardPayloadResponse convierte directamente la respuesta de la función SQL a DTO
func FromDashboardPayloadResponse(payload interface{}) DashboardResponse {
	// Convertir el payload a JSON y luego parsearlo
	jsonData, err := json.Marshal(payload)
	if err != nil {
		// Si hay error, retornar estructura vacía con valores por defecto
		return createEmptyDashboardResponse()
	}

	var response DashboardResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		// Si hay error, retornar estructura vacía con valores por defecto
		return createEmptyDashboardResponse()
	}

	// Aplicar redondeo a 3 decimales a todos los campos decimal
	response = RoundAllDecimals(response)

	return response
}

// createEmptyDashboardResponse crea una respuesta vacía con valores por defecto
func createEmptyDashboardResponse() DashboardResponse {
	return DashboardResponse{
		Metrics: Metrics{
			Sowing: SowingMetric{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Harvest: HarvestMetric{
				ProgressPct:   decimal.Zero,
				Hectares:      decimal.Zero,
				TotalHectares: decimal.Zero,
			},
			Costs: CostsMetric{
				ProgressPct: decimal.Zero,
				ExecutedUSD: decimal.Zero,
				BudgetUSD:   decimal.Zero,
			},
			InvestorContributions: InvestorContributions{
				ProgressPct: decimal.Zero,
				Breakdown:   nil,
			},
			OperatingResult: OperatingResultMetric{
				ProgressPct:   decimal.Zero,
				IncomeUSD:     decimal.Zero,
				TotalCostsUSD: decimal.Zero,
			},
		},
		ManagementBalance: ManagementBalance{
			Summary: BalanceSummary{
				IncomeUSD:              decimal.Zero,
				DirectCostsExecutedUSD: decimal.Zero,
				DirectCostsInvestedUSD: decimal.Zero,
				StockUSD:               decimal.Zero,
				RentUSD:                decimal.Zero,
				StructureUSD:           decimal.Zero,
				OperatingResultUSD:     decimal.Zero,
				OperatingResultPct:     decimal.Zero,
			},
			Breakdown: []BalanceBreakdown{},
			TotalsRow: BalanceTotals{
				ExecutedUSD: decimal.Zero,
				InvestedUSD: decimal.Zero,
				StockUSD:    decimal.Zero,
			},
		},
		CropIncidence: CropIncidence{
			Crops: []CropData{},
			Total: CropTotal{
				Hectares:          decimal.Zero,
				RotationPct:       decimal.Zero,
				CostUSDPerHectare: decimal.Zero,
			},
		},
		OperationalIndicators: OperationalIndicators{
			Cards: []OperationalCard{},
		},
	}
}

// FromDashboard convierte el dominio Dashboard a DTO
func FromDashboard(domain *domain.Dashboard) DashboardResponse {
	// Convertir el dominio a JSON y luego parsearlo
	jsonData, err := json.Marshal(domain)
	if err != nil {
		// Si hay error, retornar estructura vacía
		return DashboardResponse{}
	}

	var response DashboardResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		// Si hay error, retornar estructura vacía
		return DashboardResponse{}
	}

	// Aplicar redondeo a 3 decimales a todos los campos decimal
	response = RoundAllDecimals(response)

	return response
}

// Funciones auxiliares para crear valores decimal y punteros
func DecimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func DecimalZero() decimal.Decimal {
	return decimal.Zero
}

func StringPtr(s string) *string {
	return &s
}

// RoundAllDecimals aplica redondeo a 3 decimales a todos los campos decimal
func RoundAllDecimals(response DashboardResponse) DashboardResponse {
	// Redondear métricas
	response.Metrics.Sowing.ProgressPct = roundTo3Decimals(response.Metrics.Sowing.ProgressPct)
	response.Metrics.Sowing.Hectares = roundTo3Decimals(response.Metrics.Sowing.Hectares)
	response.Metrics.Sowing.TotalHectares = roundTo3Decimals(response.Metrics.Sowing.TotalHectares)

	response.Metrics.Harvest.ProgressPct = roundTo3Decimals(response.Metrics.Harvest.ProgressPct)
	response.Metrics.Harvest.Hectares = roundTo3Decimals(response.Metrics.Harvest.Hectares)
	response.Metrics.Harvest.TotalHectares = roundTo3Decimals(response.Metrics.Harvest.TotalHectares)

	response.Metrics.Costs.ProgressPct = roundTo3Decimals(response.Metrics.Costs.ProgressPct)
	response.Metrics.Costs.ExecutedUSD = roundTo3Decimals(response.Metrics.Costs.ExecutedUSD)
	response.Metrics.Costs.BudgetUSD = roundTo3Decimals(response.Metrics.Costs.BudgetUSD)

	response.Metrics.InvestorContributions.ProgressPct = roundTo3Decimals(response.Metrics.InvestorContributions.ProgressPct)

	response.Metrics.OperatingResult.ProgressPct = roundTo3Decimals(response.Metrics.OperatingResult.ProgressPct)
	response.Metrics.OperatingResult.IncomeUSD = roundTo3Decimals(response.Metrics.OperatingResult.IncomeUSD)
	response.Metrics.OperatingResult.TotalCostsUSD = roundTo3Decimals(response.Metrics.OperatingResult.TotalCostsUSD)

	// Redondear balance de gestión
	response.ManagementBalance.Summary.IncomeUSD = roundTo3Decimals(response.ManagementBalance.Summary.IncomeUSD)
	response.ManagementBalance.Summary.DirectCostsExecutedUSD = roundTo3Decimals(response.ManagementBalance.Summary.DirectCostsExecutedUSD)
	response.ManagementBalance.Summary.DirectCostsInvestedUSD = roundTo3Decimals(response.ManagementBalance.Summary.DirectCostsInvestedUSD)
	response.ManagementBalance.Summary.StockUSD = roundTo3Decimals(response.ManagementBalance.Summary.StockUSD)
	response.ManagementBalance.Summary.RentUSD = roundTo3Decimals(response.ManagementBalance.Summary.RentUSD)
	response.ManagementBalance.Summary.StructureUSD = roundTo3Decimals(response.ManagementBalance.Summary.StructureUSD)
	response.ManagementBalance.Summary.OperatingResultUSD = roundTo3Decimals(response.ManagementBalance.Summary.OperatingResultUSD)
	response.ManagementBalance.Summary.OperatingResultPct = roundTo3Decimals(response.ManagementBalance.Summary.OperatingResultPct)

	// Redondear breakdown - solo si StockUSD no es NULL
	for i := range response.ManagementBalance.Breakdown {
		response.ManagementBalance.Breakdown[i].ExecutedUSD = roundTo3Decimals(response.ManagementBalance.Breakdown[i].ExecutedUSD)
		response.ManagementBalance.Breakdown[i].InvestedUSD = roundTo3Decimals(response.ManagementBalance.Breakdown[i].InvestedUSD)
		if response.ManagementBalance.Breakdown[i].StockUSD != nil {
			*response.ManagementBalance.Breakdown[i].StockUSD = roundTo3Decimals(*response.ManagementBalance.Breakdown[i].StockUSD)
		}
	}

	// Redondear totals row
	response.ManagementBalance.TotalsRow.ExecutedUSD = roundTo3Decimals(response.ManagementBalance.TotalsRow.ExecutedUSD)
	response.ManagementBalance.TotalsRow.InvestedUSD = roundTo3Decimals(response.ManagementBalance.TotalsRow.InvestedUSD)
	response.ManagementBalance.TotalsRow.StockUSD = roundTo3Decimals(response.ManagementBalance.TotalsRow.StockUSD)

	// Redondear incidencia por cultivo
	for i := range response.CropIncidence.Crops {
		response.CropIncidence.Crops[i].Hectares = roundTo3Decimals(response.CropIncidence.Crops[i].Hectares)
		response.CropIncidence.Crops[i].RotationPct = roundTo3Decimals(response.CropIncidence.Crops[i].RotationPct)
		response.CropIncidence.Crops[i].CostUSDPerHa = roundTo3Decimals(response.CropIncidence.Crops[i].CostUSDPerHa)
		response.CropIncidence.Crops[i].IncidencePct = roundTo3Decimals(response.CropIncidence.Crops[i].IncidencePct)
	}

	response.CropIncidence.Total.Hectares = roundTo3Decimals(response.CropIncidence.Total.Hectares)
	response.CropIncidence.Total.RotationPct = roundTo3Decimals(response.CropIncidence.Total.RotationPct)
	response.CropIncidence.Total.CostUSDPerHectare = roundTo3Decimals(response.CropIncidence.Total.CostUSDPerHectare)

	return response
}

// ===== MAPPER FUNCTIONS =====

// ToDashboardFilter convierte un DTO de filtro a entidad de dominio
func ToDashboardFilter(dto DashboardFilterRequest) domain.DashboardFilter {
	filter := domain.DashboardFilter{}

	// Convertir slices a punteros individuales (el dominio usa punteros, no slices)
	if len(dto.CustomerIDs) > 0 {
		filter.CustomerID = &dto.CustomerIDs[0]
	}
	if len(dto.ProjectIDs) > 0 {
		filter.ProjectID = &dto.ProjectIDs[0]
	}
	if len(dto.CampaignIDs) > 0 {
		filter.CampaignID = &dto.CampaignIDs[0]
	}
	if len(dto.FieldIDs) > 0 {
		filter.FieldID = &dto.FieldIDs[0]
	}

	return filter
}

// ToDashboardData convierte un DTO de dashboard a entidad de dominio DashboardData
// (útil para casos donde se recibe un dashboard desde el exterior)
func ToDashboardData(dto DashboardRequest) *domain.DashboardData {
	return &domain.DashboardData{
		Metrics: &domain.DashboardMetrics{
			Sowing: &domain.DashboardSowing{
				ProgressPct:   dto.Metrics.Sowing.ProgressPct,
				Hectares:      dto.Metrics.Sowing.Hectares,
				TotalHectares: dto.Metrics.Sowing.TotalHectares,
			},
			Harvest: &domain.DashboardHarvest{
				ProgressPct:   dto.Metrics.Harvest.ProgressPct,
				Hectares:      dto.Metrics.Harvest.Hectares,
				TotalHectares: dto.Metrics.Harvest.TotalHectares,
			},
			Costs: &domain.DashboardCosts{
				ProgressPct: dto.Metrics.Costs.ProgressPct,
				ExecutedUSD: dto.Metrics.Costs.ExecutedUSD,
				BudgetUSD:   dto.Metrics.Costs.BudgetUSD,
			},
			InvestorContributions: &domain.DashboardInvestorContributions{
				ProgressPct: dto.Metrics.InvestorContributions.ProgressPct,
				Breakdown:   []domain.DashboardInvestorBreakdown{},
			},
			OperatingResult: &domain.DashboardOperatingResult{
				ProgressPct:   dto.Metrics.OperatingResult.ProgressPct,
				ResultUSD:     dto.Metrics.OperatingResult.IncomeUSD,
				TotalCostsUSD: dto.Metrics.OperatingResult.TotalCostsUSD,
			},
		},
		ManagementBalance: &domain.DashboardManagementBalance{
			Summary: &domain.DashboardBalanceSummary{
				IncomeUSD:              dto.ManagementBalance.Summary.IncomeUSD,
				DirectCostsExecutedUSD: dto.ManagementBalance.Summary.DirectCostsExecutedUSD,
				DirectCostsInvestedUSD: dto.ManagementBalance.Summary.DirectCostsInvestedUSD,
				StockUSD:               dto.ManagementBalance.Summary.StockUSD,
				RentUSD:                dto.ManagementBalance.Summary.RentUSD,
				StructureUSD:           dto.ManagementBalance.Summary.StructureUSD,
				OperatingResultUSD:     dto.ManagementBalance.Summary.OperatingResultUSD,
				OperatingResultPct:     dto.ManagementBalance.Summary.OperatingResultPct,
			},
			Breakdown: []domain.DashboardBalanceBreakdown{},
			TotalsRow: &domain.DashboardBalanceTotals{
				ExecutedUSD: dto.ManagementBalance.TotalsRow.ExecutedUSD,
				InvestedUSD: dto.ManagementBalance.TotalsRow.InvestedUSD,
				StockUSD:    dto.ManagementBalance.TotalsRow.StockUSD,
			},
		},
		CropIncidence: &domain.DashboardCropIncidence{
			Crops: []domain.DashboardCropBreakdown{},
			Total: &domain.DashboardCropTotal{
				Hectares:          dto.CropIncidence.Total.Hectares,
				RotationPct:       dto.CropIncidence.Total.RotationPct,
				CostUSDPerHectare: dto.CropIncidence.Total.CostUSDPerHectare,
			},
		},
		OperationalIndicators: &domain.DashboardOperationalIndicators{
			Cards: []domain.DashboardOperationalCard{},
		},
	}
}

// ===== NUEVOS MAPPER FUNCTIONS PARA DOMAIN.DashboardData =====

// FromDashboardData convierte el dominio DashboardData a DTO viejo (formato externo)
func FromDashboardData(domainData *domain.DashboardData) DashboardResponse {
	if domainData == nil {
		return createEmptyDashboardResponse()
	}

	response := DashboardResponse{}

	// Mapear métricas
	if domainData.Metrics != nil {
		response.Metrics = Metrics{
			Sowing:                mapSowingMetric(domainData.Metrics.Sowing),
			Harvest:               mapHarvestMetric(domainData.Metrics.Harvest),
			Costs:                 mapCostsMetric(domainData.Metrics.Costs),
			InvestorContributions: mapInvestorContributions(domainData.Metrics.InvestorContributions),
			OperatingResult:       mapOperatingResultMetric(domainData.Metrics.OperatingResult),
		}
	}

	// Mapear balance de gestión
	if domainData.ManagementBalance != nil {
		response.ManagementBalance = mapManagementBalance(domainData.ManagementBalance)
	}

	// Mapear incidencia de cultivos
	if domainData.CropIncidence != nil {
		response.CropIncidence = mapCropIncidence(domainData.CropIncidence)
	}

	// Mapear indicadores operativos
	if domainData.OperationalIndicators != nil {
		response.OperationalIndicators = mapOperationalIndicators(domainData.OperationalIndicators)
	}

	// Aplicar redondeo a 3 decimales
	response = RoundAllDecimals(response)

	return response
}

// mapSowingMetric mapea la métrica de siembra del dominio al DTO
func mapSowingMetric(domainSowing *domain.DashboardSowing) SowingMetric {
	if domainSowing == nil {
		return SowingMetric{
			ProgressPct:   decimal.Zero,
			Hectares:      decimal.Zero,
			TotalHectares: decimal.Zero,
		}
	}

	return SowingMetric{
		ProgressPct:   domainSowing.ProgressPct,
		Hectares:      domainSowing.Hectares,
		TotalHectares: domainSowing.TotalHectares,
	}
}

// mapHarvestMetric mapea la métrica de cosecha del dominio al DTO
func mapHarvestMetric(domainHarvest *domain.DashboardHarvest) HarvestMetric {
	if domainHarvest == nil {
		return HarvestMetric{
			ProgressPct:   decimal.Zero,
			Hectares:      decimal.Zero,
			TotalHectares: decimal.Zero,
		}
	}

	return HarvestMetric{
		ProgressPct:   domainHarvest.ProgressPct,
		Hectares:      domainHarvest.Hectares,
		TotalHectares: domainHarvest.TotalHectares,
	}
}

// mapCostsMetric mapea la métrica de costos del dominio al DTO
func mapCostsMetric(domainCosts *domain.DashboardCosts) CostsMetric {
	if domainCosts == nil {
		return CostsMetric{
			ProgressPct: decimal.Zero,
			ExecutedUSD: decimal.Zero,
			BudgetUSD:   decimal.Zero,
		}
	}

	return CostsMetric{
		ProgressPct: domainCosts.ProgressPct,
		ExecutedUSD: domainCosts.ExecutedUSD,
		BudgetUSD:   domainCosts.BudgetUSD,
	}
}

// mapInvestorContributions mapea las contribuciones de inversores del dominio al DTO
func mapInvestorContributions(domainContributions *domain.DashboardInvestorContributions) InvestorContributions {
	if domainContributions == nil {
		return InvestorContributions{
			ProgressPct: decimal.Zero,
			Breakdown:   nil,
		}
	}

	// Convertir breakdown a interface{} para mantener compatibilidad con el DTO viejo
	var breakdown interface{}
	if len(domainContributions.Breakdown) > 0 {
		breakdown = domainContributions.Breakdown
	}

	return InvestorContributions{
		ProgressPct: domainContributions.ProgressPct,
		Breakdown:   breakdown,
	}
}

// mapOperatingResultMetric mapea la métrica de resultado operativo del dominio al DTO
func mapOperatingResultMetric(domainOperating *domain.DashboardOperatingResult) OperatingResultMetric {
	if domainOperating == nil {
		return OperatingResultMetric{
			ProgressPct:   decimal.Zero,
			IncomeUSD:     decimal.Zero,
			TotalCostsUSD: decimal.Zero,
		}
	}

	return OperatingResultMetric{
		ProgressPct:   domainOperating.ProgressPct,
		IncomeUSD:     domainOperating.ResultUSD, // Mapear ResultUSD a IncomeUSD para compatibilidad
		TotalCostsUSD: domainOperating.TotalCostsUSD,
	}
}

// mapManagementBalance mapea el balance de gestión del dominio al DTO
func mapManagementBalance(domainBalance *domain.DashboardManagementBalance) ManagementBalance {
	if domainBalance == nil {
		return ManagementBalance{
			Summary:   BalanceSummary{},
			Breakdown: []BalanceBreakdown{},
			TotalsRow: BalanceTotals{},
		}
	}

	balance := ManagementBalance{}

	// Mapear summary
	if domainBalance.Summary != nil {
		balance.Summary = BalanceSummary{
			IncomeUSD:              domainBalance.Summary.IncomeUSD,
			DirectCostsExecutedUSD: domainBalance.Summary.DirectCostsExecutedUSD,
			DirectCostsInvestedUSD: domainBalance.Summary.DirectCostsInvestedUSD,
			StockUSD:               domainBalance.Summary.StockUSD,
			RentUSD:                domainBalance.Summary.RentUSD,
			StructureUSD:           domainBalance.Summary.StructureUSD,
			OperatingResultUSD:     domainBalance.Summary.OperatingResultUSD,
			OperatingResultPct:     domainBalance.Summary.OperatingResultPct,
		}
	}

	// Mapear breakdown
	balance.Breakdown = make([]BalanceBreakdown, len(domainBalance.Breakdown))
	for i, domainBreakdown := range domainBalance.Breakdown {
		balance.Breakdown[i] = BalanceBreakdown{
			Label:       domainBreakdown.Label,
			ExecutedUSD: domainBreakdown.ExecutedUSD,
			InvestedUSD: domainBreakdown.InvestedUSD,
			StockUSD:    domainBreakdown.StockUSD,
		}
	}

	// Mapear totals row
	if domainBalance.TotalsRow != nil {
		balance.TotalsRow = BalanceTotals{
			ExecutedUSD: domainBalance.TotalsRow.ExecutedUSD,
			InvestedUSD: domainBalance.TotalsRow.InvestedUSD,
			StockUSD:    domainBalance.TotalsRow.StockUSD,
		}
	}

	return balance
}

// mapCropIncidence mapea la incidencia de cultivos del dominio al DTO
func mapCropIncidence(domainCropIncidence *domain.DashboardCropIncidence) CropIncidence {
	if domainCropIncidence == nil {
		return CropIncidence{
			Crops: []CropData{},
			Total: CropTotal{},
		}
	}

	cropIncidence := CropIncidence{}

	// Mapear crops
	cropIncidence.Crops = make([]CropData, len(domainCropIncidence.Crops))
	for i, domainCrop := range domainCropIncidence.Crops {
		cropIncidence.Crops[i] = CropData{
			Name:         domainCrop.Name,
			Hectares:     domainCrop.Hectares,
			RotationPct:  domainCrop.RotationPct,
			CostUSDPerHa: domainCrop.CostUSDPerHa,
			IncidencePct: domainCrop.IncidencePct,
		}
	}

	// Mapear total
	if domainCropIncidence.Total != nil {
		cropIncidence.Total = CropTotal{
			Hectares:          domainCropIncidence.Total.Hectares,
			RotationPct:       domainCropIncidence.Total.RotationPct,
			CostUSDPerHectare: domainCropIncidence.Total.CostUSDPerHectare,
		}
	}

	return cropIncidence
}

// mapOperationalIndicators mapea los indicadores operativos del dominio al DTO
func mapOperationalIndicators(domainIndicators *domain.DashboardOperationalIndicators) OperationalIndicators {
	if domainIndicators == nil {
		return OperationalIndicators{
			Cards: []OperationalCard{},
		}
	}

	indicators := OperationalIndicators{
		Cards: make([]OperationalCard, len(domainIndicators.Cards)),
	}

	for i, domainCard := range domainIndicators.Cards {
		indicators.Cards[i] = OperationalCard{
			Key:           domainCard.Key,
			Title:         domainCard.Title,
			Date:          domainCard.Date,
			WorkorderID:   domainCard.WorkorderID,
			WorkorderCode: nil, // No existe en el dominio interno
			AuditID:       nil, // No existe en el dominio interno
			AuditCode:     nil, // No existe en el dominio interno
			Status:        nil, // No existe en el dominio interno
		}
	}

	return indicators
}
