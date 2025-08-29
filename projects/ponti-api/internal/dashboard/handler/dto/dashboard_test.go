package dto

import (
	"testing"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestRoundAllDecimals(t *testing.T) {
	// Crear una respuesta de prueba
	response := DashboardResponse{
		Metrics: Metrics{
			Sowing: SowingMetric{
				ProgressPct:   decimal.NewFromFloat(12.3456789),
				Hectares:      decimal.NewFromFloat(45.6789012),
				TotalHectares: decimal.NewFromFloat(100.1234567),
			},
			Harvest: HarvestMetric{
				ProgressPct:   decimal.NewFromFloat(23.4567890),
				Hectares:      decimal.NewFromFloat(67.8901234),
				TotalHectares: decimal.NewFromFloat(150.2345678),
			},
			Costs: CostsMetric{
				ProgressPct: decimal.NewFromFloat(34.5678901),
				ExecutedUSD: decimal.NewFromFloat(89.0123456),
				BudgetUSD:   decimal.NewFromFloat(200.3456789),
			},
			InvestorContributions: InvestorContributions{
				ProgressPct: decimal.NewFromFloat(45.6789012),
				Breakdown:   nil,
			},
			OperatingResult: OperatingResultMetric{
				ProgressPct:   decimal.NewFromFloat(56.7890123),
				IncomeUSD:     decimal.NewFromFloat(123.4567890),
				TotalCostsUSD: decimal.NewFromFloat(234.5678901),
			},
		},
	}

	// Aplicar redondeo
	result := RoundAllDecimals(response)

	// Verificar que los valores se redondeen a 3 decimales
	// Nota: los ceros finales se eliminan automáticamente
	assert.Equal(t, "12.346", result.Metrics.Sowing.ProgressPct.String())
	assert.Equal(t, "45.679", result.Metrics.Sowing.Hectares.String())
	assert.Equal(t, "100.123", result.Metrics.Sowing.TotalHectares.String())

	assert.Equal(t, "23.457", result.Metrics.Harvest.ProgressPct.String())
	assert.Equal(t, "67.89", result.Metrics.Harvest.Hectares.String()) // 67.890 se convierte en 67.89
	assert.Equal(t, "150.235", result.Metrics.Harvest.TotalHectares.String())

	assert.Equal(t, "34.568", result.Metrics.Costs.ProgressPct.String())
	assert.Equal(t, "89.012", result.Metrics.Costs.ExecutedUSD.String())
	assert.Equal(t, "200.346", result.Metrics.Costs.BudgetUSD.String())

	assert.Equal(t, "45.679", result.Metrics.InvestorContributions.ProgressPct.String())

	assert.Equal(t, "56.789", result.Metrics.OperatingResult.ProgressPct.String())
	assert.Equal(t, "123.457", result.Metrics.OperatingResult.IncomeUSD.String())
	assert.Equal(t, "234.568", result.Metrics.OperatingResult.TotalCostsUSD.String())
}

func TestCreateEmptyDashboardResponse(t *testing.T) {
	// Crear respuesta vacía
	response := createEmptyDashboardResponse()

	// Verificar que todos los campos decimal sean cero
	assert.Equal(t, decimal.Zero, response.Metrics.Sowing.ProgressPct)
	assert.Equal(t, decimal.Zero, response.Metrics.Sowing.Hectares)
	assert.Equal(t, decimal.Zero, response.Metrics.Sowing.TotalHectares)

	assert.Equal(t, decimal.Zero, response.Metrics.Harvest.ProgressPct)
	assert.Equal(t, decimal.Zero, response.Metrics.Harvest.Hectares)
	assert.Equal(t, decimal.Zero, response.Metrics.Harvest.TotalHectares)

	assert.Equal(t, decimal.Zero, response.Metrics.Costs.ProgressPct)
	assert.Equal(t, decimal.Zero, response.Metrics.Costs.ExecutedUSD)
	assert.Equal(t, decimal.Zero, response.Metrics.Costs.BudgetUSD)

	assert.Equal(t, decimal.Zero, response.Metrics.InvestorContributions.ProgressPct)
	assert.Nil(t, response.Metrics.InvestorContributions.Breakdown)

	assert.Equal(t, decimal.Zero, response.Metrics.OperatingResult.ProgressPct)
	assert.Equal(t, decimal.Zero, response.Metrics.OperatingResult.IncomeUSD)
	assert.Equal(t, decimal.Zero, response.Metrics.OperatingResult.TotalCostsUSD)

	// Verificar que los arrays estén vacíos
	assert.Empty(t, response.ManagementBalance.Breakdown)
	assert.Empty(t, response.CropIncidence.Crops)
	assert.Empty(t, response.OperationalIndicators.Cards)
}

func TestFromDashboardPayload(t *testing.T) {
	// Crear un payload de dominio con datos reales
	payload := &domain.DashboardPayload{
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
				ProgressPct:   decimal.NewFromFloat(25.5),
				Hectares:      decimal.NewFromFloat(12.75),
				TotalHectares: decimal.NewFromFloat(50.0),
			},
			Harvest: struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				Hectares      decimal.Decimal `json:"hectares"`
				TotalHectares decimal.Decimal `json:"total_hectares"`
			}{
				ProgressPct:   decimal.NewFromFloat(15.2),
				Hectares:      decimal.NewFromFloat(7.6),
				TotalHectares: decimal.NewFromFloat(50.0),
			},
			Costs: struct {
				ProgressPct decimal.Decimal `json:"progress_pct"`
				ExecutedUSD decimal.Decimal `json:"executed_usd"`
				BudgetUSD   decimal.Decimal `json:"budget_usd"`
			}{
				ProgressPct: decimal.NewFromFloat(45.8),
				ExecutedUSD: decimal.NewFromFloat(4580.0),
				BudgetUSD:   decimal.NewFromFloat(10000.0),
			},
			InvestorContributions: struct {
				ProgressPct decimal.Decimal `json:"progress_pct"`
				Breakdown   interface{}     `json:"breakdown"`
			}{
				ProgressPct: decimal.NewFromFloat(60.0),
				Breakdown:   []interface{}{},
			},
			OperatingResult: struct {
				ProgressPct   decimal.Decimal `json:"progress_pct"`
				IncomeUSD     decimal.Decimal `json:"income_usd"`
				TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
			}{
				ProgressPct:   decimal.NewFromFloat(12.5),
				IncomeUSD:     decimal.NewFromFloat(5000.0),
				TotalCostsUSD: decimal.NewFromFloat(4000.0),
			},
		},
	}

	// Convertir el payload de dominio a DTO
	response := FromDashboardPayload(payload)

	// Verificar que los datos se hayan convertido correctamente
	assert.Equal(t, "25.5", response.Metrics.Sowing.ProgressPct.String())
	assert.Equal(t, "12.75", response.Metrics.Sowing.Hectares.String())
	assert.Equal(t, "50", response.Metrics.Sowing.TotalHectares.String())

	assert.Equal(t, "15.2", response.Metrics.Harvest.ProgressPct.String())
	assert.Equal(t, "7.6", response.Metrics.Harvest.Hectares.String())
	assert.Equal(t, "50", response.Metrics.Harvest.TotalHectares.String())

	assert.Equal(t, "45.8", response.Metrics.Costs.ProgressPct.String())
	assert.Equal(t, "4580", response.Metrics.Costs.ExecutedUSD.String())
	assert.Equal(t, "10000", response.Metrics.Costs.BudgetUSD.String())

	assert.Equal(t, "60", response.Metrics.InvestorContributions.ProgressPct.String())
	assert.NotNil(t, response.Metrics.InvestorContributions.Breakdown)

	assert.Equal(t, "12.5", response.Metrics.OperatingResult.ProgressPct.String())
	assert.Equal(t, "5000", response.Metrics.OperatingResult.IncomeUSD.String())
	assert.Equal(t, "4000", response.Metrics.OperatingResult.TotalCostsUSD.String())

	// Verificar que el redondeo se haya aplicado
	assert.Equal(t, "25.5", response.Metrics.Sowing.ProgressPct.String())
	assert.Equal(t, "12.75", response.Metrics.Sowing.Hectares.String())
}
