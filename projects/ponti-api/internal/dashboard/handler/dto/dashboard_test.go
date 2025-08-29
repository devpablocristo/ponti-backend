package dto

import (
	"testing"

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
