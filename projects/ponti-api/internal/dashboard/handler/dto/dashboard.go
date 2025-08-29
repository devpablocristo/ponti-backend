package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

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

	// Redondear breakdown
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
