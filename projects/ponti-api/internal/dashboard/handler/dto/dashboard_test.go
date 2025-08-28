package dto

import (
	"testing"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    decimal.NullDecimal
		expected *decimal.Decimal
	}{
		{
			name:     "valid decimal returns pointer",
			input:    decimal.NullDecimal{Decimal: decimal.NewFromFloat(123.45), Valid: true},
			expected: decimalPtr(decimal.NewFromFloat(123.45)),
		},
		{
			name:     "invalid decimal returns nil",
			input:    decimal.NullDecimal{Decimal: decimal.Zero, Valid: false},
			expected: nil,
		},
		{
			name:     "zero decimal with valid true returns pointer",
			input:    decimal.NullDecimal{Decimal: decimal.Zero, Valid: true},
			expected: decimalPtr(decimal.Zero),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toPtr(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.True(t, result.Equal(*tt.expected))
			}
		})
	}
}

func TestToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected decimal.Decimal
	}{
		{
			name:     "float64 conversion",
			input:    123.45,
			expected: decimal.NewFromFloat(123.45),
		},
		{
			name:     "float32 conversion",
			input:    float32(67.89),
			expected: decimal.NewFromFloat32(67.89),
		},
		{
			name:     "int conversion",
			input:    42,
			expected: decimal.NewFromInt(42),
		},
		{
			name:     "int64 conversion",
			input:    int64(999),
			expected: decimal.NewFromInt(999),
		},
		{
			name:     "string conversion valid decimal",
			input:    "123.456",
			expected: decimal.RequireFromString("123.456"),
		},
		{
			name:     "string conversion invalid decimal returns zero",
			input:    "invalid",
			expected: decimal.Zero,
		},
		{
			name:     "decimal.Decimal passthrough",
			input:    decimal.NewFromFloat(100.25),
			expected: decimal.NewFromFloat(100.25),
		},
		{
			name:     "nil input returns zero",
			input:    nil,
			expected: decimal.Zero,
		},
		{
			name:     "bool input returns zero",
			input:    true,
			expected: decimal.Zero,
		},
		{
			name:     "slice input returns zero",
			input:    []int{1, 2, 3},
			expected: decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDecimal(tt.input)
			assert.True(t, result.Equal(tt.expected), "Expected %s, got %s", tt.expected.String(), result.String())
		})
	}
}

func TestFromDomain(t *testing.T) {
	// Test data setup
	now := time.Now()
	zeroDecimal := decimal.Zero
	oneDecimal := decimal.NewFromInt(1)
	hundredDecimal := decimal.NewFromFloat(100.0)
	thousandDecimal := decimal.NewFromFloat(1000.0)

	validNullDecimal := decimal.NullDecimal{Decimal: decimal.NewFromFloat(50.0), Valid: true}
	invalidNullDecimal := decimal.NullDecimal{Decimal: decimal.Zero, Valid: false}

	tests := []struct {
		name           string
		input          domain.DashboardRow
		expectedFields func() DashboardRow
		description    string
	}{
		{
			name: "complete dashboard with all fields populated",
			input: domain.DashboardRow{
				TotalHectares:            hundredDecimal,
				SowedArea:                decimal.NewFromFloat(75.0),
				HarvestedArea:            decimal.NewFromFloat(50.0),
				SowingProgressPct:        validNullDecimal,
				HarvestProgressPct:       validNullDecimal,
				ExecutedCosts:            thousandDecimal,
				BudgetCosts:              decimal.NullDecimal{Decimal: decimal.NewFromFloat(1200.0), Valid: true},
				IncomeNet:                decimal.NewFromFloat(2000.0),
				TotalCosts:               thousandDecimal,
				ContributionDetails:      stringPtr("Investor A: 60%, Investor B: 40%"),
				CostsProgressPct:         validNullDecimal,
				OperatingResultPct:       validNullDecimal,
				CropsBreakdown:           stringPtr(`{"soja":{"hectares":50.0,"total_cost":500.0,"cost_per_ha":10.0,"rotation_pct":50.0},"maiz":{"hectares":50.0,"total_cost":500.0,"cost_per_ha":10.0,"rotation_pct":50.0}}`),
				CropsTotalHectares:       hundredDecimal,
				CropsTotalRotationPct:    decimal.NewFromFloat(100.0),
				CropsTotalCostPerHectare: decimal.NewFromFloat(10.0),
				YieldPerHectare:          decimal.NullDecimal{Decimal: decimal.NewFromFloat(8.5), Valid: true},
				TotalCostPerHectare:      decimal.NullDecimal{Decimal: decimal.NewFromFloat(12.0), Valid: true},
				FirstOrderDate:           &now,
				FirstOrderNumber:         stringPtr("WO-001"),
				LastOrderDate:            &now,
				LastOrderNumber:          stringPtr("WO-100"),
				LastStockCountDate:       &now,
				MgmtIncomeUSD:            decimal.NewFromFloat(2000.0),
				MgmtTotalCostsUSD:        thousandDecimal,
				MgmtOperatingResultUSD:   decimal.NewFromFloat(1000.0),
				MgmtOperatingResultPct:   validNullDecimal,
				InvestedCostUSD:          decimal.NullDecimal{Decimal: decimal.NewFromFloat(800.0), Valid: true},
				StockUSD:                 decimal.NullDecimal{Decimal: decimal.NewFromFloat(200.0), Valid: true},
			},
			expectedFields: func() DashboardRow {
				return DashboardRow{
					Metrics: struct {
						Sowing struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						} `json:"sowing"`
						Harvest struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						} `json:"harvest"`
						Costs struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							Executed    decimal.Decimal  `json:"executed"`
							Budget      *decimal.Decimal `json:"budget,omitempty"`
						} `json:"costs"`
						InvestorContributions struct {
							Details *string `json:"details,omitempty"`
						} `json:"investor_contributions"`
						OperatingResult struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							IncomeNet   decimal.Decimal  `json:"income_net"`
							TotalCosts  decimal.Decimal  `json:"total_costs"`
						} `json:"operating_result"`
					}{
						Sowing: struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						}{
							ProgressPct:   decimalPtr(decimal.NewFromFloat(50.0)),
							Hectares:      decimal.NewFromFloat(75.0),
							TotalHectares: hundredDecimal,
						},
						Harvest: struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						}{
							ProgressPct:   decimalPtr(decimal.NewFromFloat(50.0)),
							Hectares:      decimal.NewFromFloat(50.0),
							TotalHectares: hundredDecimal,
						},
						Costs: struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							Executed    decimal.Decimal  `json:"executed"`
							Budget      *decimal.Decimal `json:"budget,omitempty"`
						}{
							ProgressPct: decimalPtr(decimal.NewFromFloat(50.0)),
							Executed:    thousandDecimal,
							Budget:      decimalPtr(decimal.NewFromFloat(1200.0)),
						},
						InvestorContributions: struct {
							Details *string `json:"details,omitempty"`
						}{
							Details: stringPtr("Investor A: 60%, Investor B: 40%"),
						},
						OperatingResult: struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							IncomeNet   decimal.Decimal  `json:"income_net"`
							TotalCosts  decimal.Decimal  `json:"total_costs"`
						}{
							ProgressPct: decimalPtr(decimal.NewFromFloat(50.0)),
							IncomeNet:   decimal.NewFromFloat(2000.0),
							TotalCosts:  thousandDecimal,
						},
					},
					CropIncidence: struct {
						Crops []CropDetail `json:"crops"`
						Total struct {
							Hectares       decimal.Decimal `json:"hectares"`
							RotationPct    decimal.Decimal `json:"rotation_pct"`
							CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
						} `json:"total"`
					}{
						Crops: []CropDetail{
							{
								Name:        "soja",
								Hectares:    decimal.NewFromFloat(50.0),
								TotalCost:   decimal.NewFromFloat(500.0),
								CostPerHa:   decimal.NewFromFloat(10.0),
								RotationPct: decimal.NewFromFloat(50.0),
							},
							{
								Name:        "maiz",
								Hectares:    decimal.NewFromFloat(50.0),
								TotalCost:   decimal.NewFromFloat(500.0),
								CostPerHa:   decimal.NewFromFloat(10.0),
								RotationPct: decimal.NewFromFloat(50.0),
							},
						},
						Total: struct {
							Hectares       decimal.Decimal `json:"hectares"`
							RotationPct    decimal.Decimal `json:"rotation_pct"`
							CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
						}{
							Hectares:       hundredDecimal,
							RotationPct:    decimal.NewFromFloat(100.0),
							CostPerHectare: decimal.NewFromFloat(10.0),
						},
					},
					Performance: struct {
						YieldPerHectare     *decimal.Decimal `json:"yield_per_hectare,omitempty"`
						TotalCostPerHectare *decimal.Decimal `json:"total_cost_per_hectare,omitempty"`
					}{
						YieldPerHectare:     decimalPtr(decimal.NewFromFloat(8.5)),
						TotalCostPerHectare: decimalPtr(decimal.NewFromFloat(12.0)),
					},
					OperationalIndicators: struct {
						FirstOrderDate     *time.Time `json:"first_order_date,omitempty"`
						FirstOrderNumber   *string    `json:"first_order_number,omitempty"`
						LastOrderDate      *time.Time `json:"last_order_date,omitempty"`
						LastOrderNumber    *string    `json:"last_order_number,omitempty"`
						LastStockCountDate *time.Time `json:"last_stock_count_date,omitempty"`
					}{
						FirstOrderDate:     &now,
						FirstOrderNumber:   stringPtr("WO-001"),
						LastOrderDate:      &now,
						LastOrderNumber:    stringPtr("WO-100"),
						LastStockCountDate: &now,
					},
					ManagementBalance: struct {
						IncomeUSD          decimal.Decimal  `json:"income_usd"`
						TotalCostsUSD      decimal.Decimal  `json:"total_costs_usd"`
						OperatingResultUSD decimal.Decimal  `json:"operating_result_usd"`
						OperatingResultPct *decimal.Decimal `json:"operating_result_pct,omitempty"`
						InvestedCostUSD    *decimal.Decimal `json:"invested_cost_usd,omitempty"`
						StockUSD           *decimal.Decimal `json:"stock_usd,omitempty"`
					}{
						IncomeUSD:          decimal.NewFromFloat(2000.0),
						TotalCostsUSD:      thousandDecimal,
						OperatingResultUSD: decimal.NewFromFloat(1000.0),
						OperatingResultPct: decimalPtr(decimal.NewFromFloat(50.0)),
						InvestedCostUSD:    decimalPtr(decimal.NewFromFloat(800.0)),
						StockUSD:           decimalPtr(decimal.NewFromFloat(200.0)),
					},
				}
			},
			description: "should convert complete domain object with all fields populated",
		},
		{
			name: "dashboard with null values and nil fields",
			input: domain.DashboardRow{
				TotalHectares:            zeroDecimal,
				SowedArea:                zeroDecimal,
				HarvestedArea:            zeroDecimal,
				SowingProgressPct:        invalidNullDecimal,
				HarvestProgressPct:       invalidNullDecimal,
				ExecutedCosts:            zeroDecimal,
				BudgetCosts:              invalidNullDecimal,
				IncomeNet:                zeroDecimal,
				TotalCosts:               zeroDecimal,
				ContributionDetails:      nil,
				CostsProgressPct:         invalidNullDecimal,
				OperatingResultPct:       invalidNullDecimal,
				CropsBreakdown:           nil,
				CropsTotalHectares:       zeroDecimal,
				CropsTotalRotationPct:    zeroDecimal,
				CropsTotalCostPerHectare: zeroDecimal,
				YieldPerHectare:          invalidNullDecimal,
				TotalCostPerHectare:      invalidNullDecimal,
				FirstOrderDate:           nil,
				FirstOrderNumber:         nil,
				LastOrderDate:            nil,
				LastOrderNumber:          nil,
				LastStockCountDate:       nil,
				MgmtIncomeUSD:            zeroDecimal,
				MgmtTotalCostsUSD:        zeroDecimal,
				MgmtOperatingResultUSD:   zeroDecimal,
				MgmtOperatingResultPct:   invalidNullDecimal,
				InvestedCostUSD:          invalidNullDecimal,
				StockUSD:                 invalidNullDecimal,
			},
			expectedFields: func() DashboardRow {
				return DashboardRow{
					Metrics: struct {
						Sowing struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						} `json:"sowing"`
						Harvest struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						} `json:"harvest"`
						Costs struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							Executed    decimal.Decimal  `json:"executed"`
							Budget      *decimal.Decimal `json:"budget,omitempty"`
						} `json:"costs"`
						InvestorContributions struct {
							Details *string `json:"details,omitempty"`
						} `json:"investor_contributions"`
						OperatingResult struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							IncomeNet   decimal.Decimal  `json:"income_net"`
							TotalCosts  decimal.Decimal  `json:"total_costs"`
						} `json:"operating_result"`
					}{
						Sowing: struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						}{
							ProgressPct:   nil,
							Hectares:      zeroDecimal,
							TotalHectares: zeroDecimal,
						},
						Harvest: struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						}{
							ProgressPct:   nil,
							Hectares:      zeroDecimal,
							TotalHectares: zeroDecimal,
						},
						Costs: struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							Executed    decimal.Decimal  `json:"executed"`
							Budget      *decimal.Decimal `json:"budget,omitempty"`
						}{
							ProgressPct: nil,
							Executed:    zeroDecimal,
							Budget:      nil,
						},
						InvestorContributions: struct {
							Details *string `json:"details,omitempty"`
						}{
							Details: nil,
						},
						OperatingResult: struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							IncomeNet   decimal.Decimal  `json:"income_net"`
							TotalCosts  decimal.Decimal  `json:"total_costs"`
						}{
							ProgressPct: nil,
							IncomeNet:   zeroDecimal,
							TotalCosts:  zeroDecimal,
						},
					},
					CropIncidence: struct {
						Crops []CropDetail `json:"crops"`
						Total struct {
							Hectares       decimal.Decimal `json:"hectares"`
							RotationPct    decimal.Decimal `json:"rotation_pct"`
							CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
						} `json:"total"`
					}{
						Crops: []CropDetail{},
						Total: struct {
							Hectares       decimal.Decimal `json:"hectares"`
							RotationPct    decimal.Decimal `json:"rotation_pct"`
							CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
						}{
							Hectares:       zeroDecimal,
							RotationPct:    zeroDecimal,
							CostPerHectare: zeroDecimal,
						},
					},
					Performance: struct {
						YieldPerHectare     *decimal.Decimal `json:"yield_per_hectare,omitempty"`
						TotalCostPerHectare *decimal.Decimal `json:"total_cost_per_hectare,omitempty"`
					}{
						YieldPerHectare:     nil,
						TotalCostPerHectare: nil,
					},
					OperationalIndicators: struct {
						FirstOrderDate     *time.Time `json:"first_order_date,omitempty"`
						FirstOrderNumber   *string    `json:"first_order_number,omitempty"`
						LastOrderDate      *time.Time `json:"last_order_date,omitempty"`
						LastOrderNumber    *string    `json:"last_order_number,omitempty"`
						LastStockCountDate *time.Time `json:"last_stock_count_date,omitempty"`
					}{
						FirstOrderDate:     nil,
						FirstOrderNumber:   nil,
						LastOrderDate:      nil,
						LastOrderNumber:    nil,
						LastStockCountDate: nil,
					},
					ManagementBalance: struct {
						IncomeUSD          decimal.Decimal  `json:"income_usd"`
						TotalCostsUSD      decimal.Decimal  `json:"total_costs_usd"`
						OperatingResultUSD decimal.Decimal  `json:"operating_result_usd"`
						OperatingResultPct *decimal.Decimal `json:"operating_result_pct,omitempty"`
						InvestedCostUSD    *decimal.Decimal `json:"invested_cost_usd,omitempty"`
						StockUSD           *decimal.Decimal `json:"stock_usd,omitempty"`
					}{
						IncomeUSD:          zeroDecimal,
						TotalCostsUSD:      zeroDecimal,
						OperatingResultUSD: zeroDecimal,
						OperatingResultPct: nil,
						InvestedCostUSD:    nil,
						StockUSD:           nil,
					},
				}
			},
			description: "should handle null values and nil fields correctly",
		},
		{
			name: "dashboard with invalid JSON in CropsBreakdown",
			input: domain.DashboardRow{
				TotalHectares:            oneDecimal,
				SowedArea:                oneDecimal,
				HarvestedArea:            oneDecimal,
				SowingProgressPct:        validNullDecimal,
				HarvestProgressPct:       validNullDecimal,
				ExecutedCosts:            oneDecimal,
				BudgetCosts:              validNullDecimal,
				IncomeNet:                oneDecimal,
				TotalCosts:               oneDecimal,
				ContributionDetails:      stringPtr("Test"),
				CostsProgressPct:         validNullDecimal,
				OperatingResultPct:       validNullDecimal,
				CropsBreakdown:           stringPtr(`{"invalid json`),
				CropsTotalHectares:       oneDecimal,
				CropsTotalRotationPct:    oneDecimal,
				CropsTotalCostPerHectare: oneDecimal,
				YieldPerHectare:          validNullDecimal,
				TotalCostPerHectare:      validNullDecimal,
				FirstOrderDate:           &now,
				FirstOrderNumber:         stringPtr("WO-001"),
				LastOrderDate:            &now,
				LastOrderNumber:          stringPtr("WO-100"),
				LastStockCountDate:       &now,
				MgmtIncomeUSD:            oneDecimal,
				MgmtTotalCostsUSD:        oneDecimal,
				MgmtOperatingResultUSD:   oneDecimal,
				MgmtOperatingResultPct:   validNullDecimal,
				InvestedCostUSD:          validNullDecimal,
				StockUSD:                 validNullDecimal,
			},
			expectedFields: func() DashboardRow {
				return DashboardRow{
					Metrics: struct {
						Sowing struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						} `json:"sowing"`
						Harvest struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						} `json:"harvest"`
						Costs struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							Executed    decimal.Decimal  `json:"executed"`
							Budget      *decimal.Decimal `json:"budget,omitempty"`
						} `json:"costs"`
						InvestorContributions struct {
							Details *string `json:"details,omitempty"`
						} `json:"investor_contributions"`
						OperatingResult struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							IncomeNet   decimal.Decimal  `json:"income_net"`
							TotalCosts  decimal.Decimal  `json:"total_costs"`
						} `json:"operating_result"`
					}{
						Sowing: struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						}{
							ProgressPct:   decimalPtr(decimal.NewFromFloat(50.0)),
							Hectares:      oneDecimal,
							TotalHectares: oneDecimal,
						},
						Harvest: struct {
							ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
							Hectares      decimal.Decimal  `json:"hectares"`
							TotalHectares decimal.Decimal  `json:"total_hectares"`
						}{
							ProgressPct:   decimalPtr(decimal.NewFromFloat(50.0)),
							Hectares:      oneDecimal,
							TotalHectares: oneDecimal,
						},
						Costs: struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							Executed    decimal.Decimal  `json:"executed"`
							Budget      *decimal.Decimal `json:"budget,omitempty"`
						}{
							ProgressPct: decimalPtr(decimal.NewFromFloat(50.0)),
							Executed:    oneDecimal,
							Budget:      decimalPtr(decimal.NewFromFloat(50.0)),
						},
						InvestorContributions: struct {
							Details *string `json:"details,omitempty"`
						}{
							Details: stringPtr("Test"),
						},
						OperatingResult: struct {
							ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
							IncomeNet   decimal.Decimal  `json:"income_net"`
							TotalCosts  decimal.Decimal  `json:"total_costs"`
						}{
							ProgressPct: decimalPtr(decimal.NewFromFloat(50.0)),
							IncomeNet:   oneDecimal,
							TotalCosts:  oneDecimal,
						},
					},
					CropIncidence: struct {
						Crops []CropDetail `json:"crops"`
						Total struct {
							Hectares       decimal.Decimal `json:"hectares"`
							RotationPct    decimal.Decimal `json:"rotation_pct"`
							CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
						} `json:"total"`
					}{
						Crops: []CropDetail{},
						Total: struct {
							Hectares       decimal.Decimal `json:"hectares"`
							RotationPct    decimal.Decimal `json:"rotation_pct"`
							CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
						}{
							Hectares:       oneDecimal,
							RotationPct:    oneDecimal,
							CostPerHectare: oneDecimal,
						},
					},
					Performance: struct {
						YieldPerHectare     *decimal.Decimal `json:"yield_per_hectare,omitempty"`
						TotalCostPerHectare *decimal.Decimal `json:"total_cost_per_hectare,omitempty"`
					}{
						YieldPerHectare:     decimalPtr(decimal.NewFromFloat(50.0)),
						TotalCostPerHectare: decimalPtr(decimal.NewFromFloat(50.0)),
					},
					OperationalIndicators: struct {
						FirstOrderDate     *time.Time `json:"first_order_date,omitempty"`
						FirstOrderNumber   *string    `json:"first_order_number,omitempty"`
						LastOrderDate      *time.Time `json:"last_order_date,omitempty"`
						LastOrderNumber    *string    `json:"last_order_number,omitempty"`
						LastStockCountDate *time.Time `json:"last_stock_count_date,omitempty"`
					}{
						FirstOrderDate:     &now,
						FirstOrderNumber:   stringPtr("WO-001"),
						LastOrderDate:      &now,
						LastOrderNumber:    stringPtr("WO-100"),
						LastStockCountDate: &now,
					},
					ManagementBalance: struct {
						IncomeUSD          decimal.Decimal  `json:"income_usd"`
						TotalCostsUSD      decimal.Decimal  `json:"total_costs_usd"`
						OperatingResultUSD decimal.Decimal  `json:"operating_result_usd"`
						OperatingResultPct *decimal.Decimal `json:"operating_result_pct,omitempty"`
						InvestedCostUSD    *decimal.Decimal `json:"invested_cost_usd,omitempty"`
						StockUSD           *decimal.Decimal `json:"stock_usd,omitempty"`
					}{
						IncomeUSD:          oneDecimal,
						TotalCostsUSD:      oneDecimal,
						OperatingResultUSD: oneDecimal,
						OperatingResultPct: decimalPtr(decimal.NewFromFloat(50.0)),
						InvestedCostUSD:    decimalPtr(decimal.NewFromFloat(50.0)),
						StockUSD:           decimalPtr(decimal.NewFromFloat(50.0)),
					},
				}
			},
			description: "should handle invalid JSON in CropsBreakdown gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromDomain(tt.input)
			expected := tt.expectedFields()

			// Test Metrics
			assert.Equal(t, expected.Metrics.Sowing.ProgressPct, result.Metrics.Sowing.ProgressPct)
			assert.True(t, expected.Metrics.Sowing.Hectares.Equal(result.Metrics.Sowing.Hectares))
			assert.True(t, expected.Metrics.Sowing.TotalHectares.Equal(result.Metrics.Sowing.TotalHectares))

			assert.Equal(t, expected.Metrics.Harvest.ProgressPct, result.Metrics.Harvest.ProgressPct)
			assert.True(t, expected.Metrics.Harvest.Hectares.Equal(result.Metrics.Harvest.Hectares))
			assert.True(t, expected.Metrics.Harvest.TotalHectares.Equal(result.Metrics.Harvest.TotalHectares))

			assert.Equal(t, expected.Metrics.Costs.ProgressPct, result.Metrics.Costs.ProgressPct)
			assert.True(t, expected.Metrics.Costs.Executed.Equal(result.Metrics.Costs.Executed))
			assert.Equal(t, expected.Metrics.Costs.Budget, result.Metrics.Costs.Budget)

			assert.Equal(t, expected.Metrics.InvestorContributions.Details, result.Metrics.InvestorContributions.Details)

			assert.Equal(t, expected.Metrics.OperatingResult.ProgressPct, result.Metrics.OperatingResult.ProgressPct)
			assert.True(t, expected.Metrics.OperatingResult.IncomeNet.Equal(result.Metrics.OperatingResult.IncomeNet))
			assert.True(t, expected.Metrics.OperatingResult.TotalCosts.Equal(result.Metrics.OperatingResult.TotalCosts))

			// Test CropIncidence
			assert.Equal(t, len(expected.CropIncidence.Crops), len(result.CropIncidence.Crops))
			if len(expected.CropIncidence.Crops) > 0 {
				for i, expectedCrop := range expected.CropIncidence.Crops {
					assert.Equal(t, expectedCrop.Name, result.CropIncidence.Crops[i].Name)
					assert.True(t, expectedCrop.Hectares.Equal(result.CropIncidence.Crops[i].Hectares))
					assert.True(t, expectedCrop.TotalCost.Equal(result.CropIncidence.Crops[i].TotalCost))
					assert.True(t, expectedCrop.CostPerHa.Equal(result.CropIncidence.Crops[i].CostPerHa))
					assert.True(t, expectedCrop.RotationPct.Equal(result.CropIncidence.Crops[i].RotationPct))
				}
			}

			assert.True(t, expected.CropIncidence.Total.Hectares.Equal(result.CropIncidence.Total.Hectares))
			assert.True(t, expected.CropIncidence.Total.RotationPct.Equal(result.CropIncidence.Total.RotationPct))
			assert.True(t, expected.CropIncidence.Total.CostPerHectare.Equal(result.CropIncidence.Total.CostPerHectare))

			// Test Performance
			if expected.Performance.YieldPerHectare != nil {
				assert.True(t, expected.Performance.YieldPerHectare.Equal(*result.Performance.YieldPerHectare))
			} else {
				assert.Nil(t, result.Performance.YieldPerHectare)
			}

			if expected.Performance.TotalCostPerHectare != nil {
				assert.True(t, expected.Performance.TotalCostPerHectare.Equal(*result.Performance.TotalCostPerHectare))
			} else {
				assert.Nil(t, result.Performance.TotalCostPerHectare)
			}

			// Test OperationalIndicators
			assert.Equal(t, expected.OperationalIndicators.FirstOrderDate, result.OperationalIndicators.FirstOrderDate)
			assert.Equal(t, expected.OperationalIndicators.FirstOrderNumber, result.OperationalIndicators.FirstOrderNumber)
			assert.Equal(t, expected.OperationalIndicators.LastOrderDate, result.OperationalIndicators.LastOrderDate)
			assert.Equal(t, expected.OperationalIndicators.LastOrderNumber, result.OperationalIndicators.LastOrderNumber)
			assert.Equal(t, expected.OperationalIndicators.LastStockCountDate, result.OperationalIndicators.LastStockCountDate)

			// Test ManagementBalance
			assert.True(t, expected.ManagementBalance.IncomeUSD.Equal(result.ManagementBalance.IncomeUSD))
			assert.True(t, expected.ManagementBalance.TotalCostsUSD.Equal(result.ManagementBalance.TotalCostsUSD))
			assert.True(t, expected.ManagementBalance.OperatingResultUSD.Equal(result.ManagementBalance.OperatingResultUSD))

			if expected.ManagementBalance.OperatingResultPct != nil {
				assert.True(t, expected.ManagementBalance.OperatingResultPct.Equal(*result.ManagementBalance.OperatingResultPct))
			} else {
				assert.Nil(t, result.ManagementBalance.OperatingResultPct)
			}

			if expected.ManagementBalance.InvestedCostUSD != nil {
				assert.True(t, expected.ManagementBalance.InvestedCostUSD.Equal(*result.ManagementBalance.InvestedCostUSD))
			} else {
				assert.Nil(t, result.ManagementBalance.InvestedCostUSD)
			}

			if expected.ManagementBalance.StockUSD != nil {
				assert.True(t, expected.ManagementBalance.StockUSD.Equal(*result.ManagementBalance.StockUSD))
			} else {
				assert.Nil(t, result.ManagementBalance.StockUSD)
			}
		})
	}
}

func TestCropDetailStruct(t *testing.T) {
	crop := CropDetail{
		Name:        "Test Crop",
		Hectares:    decimal.NewFromFloat(100.5),
		TotalCost:   decimal.NewFromFloat(50000.75),
		CostPerHa:   decimal.NewFromFloat(497.52),
		RotationPct: decimal.NewFromFloat(25.0),
	}

	assert.Equal(t, "Test Crop", crop.Name)
	assert.True(t, decimal.NewFromFloat(100.5).Equal(crop.Hectares))
	assert.True(t, decimal.NewFromFloat(50000.75).Equal(crop.TotalCost))
	assert.True(t, decimal.NewFromFloat(497.52).Equal(crop.CostPerHa))
	assert.True(t, decimal.NewFromFloat(25.0).Equal(crop.RotationPct))
}

func TestDashboardRowStruct(t *testing.T) {
	// Test that the struct can be instantiated without errors
	dashboard := DashboardRow{}

	// Test that all fields are accessible
	assert.NotNil(t, dashboard.Metrics)
	assert.NotNil(t, dashboard.CropIncidence)
	assert.NotNil(t, dashboard.Performance)
	assert.NotNil(t, dashboard.OperationalIndicators)
	assert.NotNil(t, dashboard.ManagementBalance)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create decimal pointers
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
