package dto

import (
	"encoding/json"
	"testing"
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetStocksResponse_MarshalJSON_Rounding(t *testing.T) {
	// Crear una respuesta con valores decimales que necesiten redondeo
	response := GetStocksResponse{
		Stocks:         []GetStockSummary{},           // Lista vacía para este test
		NetTotalUSD:    decimal.NewFromFloat(1234.56), // Debería redondearse a 1235
		TotalLiters:    decimal.NewFromFloat(567.89),  // Debería redondearse a 568
		TotalKilograms: decimal.NewFromFloat(901.23),  // Debería redondearse a 901
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	// Deserializar para verificar los valores
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	assert.NoError(t, err)

	// Verificar que los valores estén redondeados al entero más próximo
	assert.Equal(t, "1235", result["net_total_usd"])  // 1234.56 -> 1235
	assert.Equal(t, "568", result["total_liters"])    // 567.89 -> 568
	assert.Equal(t, "901", result["total_kilograms"]) // 901.23 -> 901
}

func TestGetStocksResponse_MarshalJSON_RoundingWithDecimals(t *testing.T) {
	// Probar casos específicos de redondeo
	tests := []struct {
		name           string
		netTotalUSD    float64
		totalLiters    float64
		totalKilograms float64
		expectedUSD    string
		expectedLiters string
		expectedKG     string
	}{
		{
			name:           "Redondeo hacia arriba",
			netTotalUSD:    1.5,
			totalLiters:    2.5,
			totalKilograms: 3.5,
			expectedUSD:    "2",
			expectedLiters: "3",
			expectedKG:     "4",
		},
		{
			name:           "Redondeo hacia abajo",
			netTotalUSD:    1.4,
			totalLiters:    2.4,
			totalKilograms: 3.4,
			expectedUSD:    "1",
			expectedLiters: "2",
			expectedKG:     "3",
		},
		{
			name:           "Valor entero exacto",
			netTotalUSD:    100.0,
			totalLiters:    200.0,
			totalKilograms: 300.0,
			expectedUSD:    "100",
			expectedLiters: "200",
			expectedKG:     "300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := GetStocksResponse{
				Stocks:         []GetStockSummary{},
				NetTotalUSD:    decimal.NewFromFloat(tt.netTotalUSD),
				TotalLiters:    decimal.NewFromFloat(tt.totalLiters),
				TotalKilograms: decimal.NewFromFloat(tt.totalKilograms),
			}

			jsonData, err := json.Marshal(response)
			assert.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(jsonData, &result)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedUSD, result["net_total_usd"])
			assert.Equal(t, tt.expectedLiters, result["total_liters"])
			assert.Equal(t, tt.expectedKG, result["total_kilograms"])
		})
	}
}

func TestFromDomain_MarshalJSON_NullStockDifferenceWhenNoRealCount(t *testing.T) {
	item := FromDomain(&stockdomain.Stock{
		Supply: &supplydomain.Supply{
			Name:         "Urea",
			CategoryName: "Fertilizantes",
			UnitID:       2,
			UnitName:     "Kg",
			Price:        decimal.NewFromInt(10),
		},
		RealStockUnits:    decimal.Zero,
		Consumed:          decimal.NewFromInt(100),
		HasRealStockCount: false,
	})

	payload, err := json.Marshal(item)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	assert.NoError(t, err)

	assert.Equal(t, "-100.00", result["stock_units"])
	assert.Equal(t, nil, result["stock_difference"])
}

func TestFromDomain_MarshalJSON_IncludesStockDifferenceWhenRealCountExists(t *testing.T) {
	item := FromDomain(&stockdomain.Stock{
		Supply: &supplydomain.Supply{
			Name:         "Urea",
			CategoryName: "Fertilizantes",
			UnitID:       2,
			UnitName:     "Kg",
			Price:        decimal.NewFromInt(10),
		},
		RealStockUnits:    decimal.NewFromInt(40),
		Consumed:          decimal.NewFromInt(2),
		SupplyMovements:   []supplydomain.SupplyMovement{{IsEntry: true, Quantity: decimal.NewFromInt(40)}},
		HasRealStockCount: true,
	})

	payload, err := json.Marshal(item)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	assert.NoError(t, err)

	assert.Equal(t, "38.00", result["stock_units"])
	assert.Equal(t, "2.00", result["stock_difference"])
}

func TestFromDomain_MarshalJSON_IncludesUpdatedAtWhenPresent(t *testing.T) {
	updatedAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	item := FromDomain(&stockdomain.Stock{
		Supply: &supplydomain.Supply{
			Name:         "Urea",
			CategoryName: "Fertilizantes",
			UnitID:       2,
			UnitName:     "Kg",
			Price:        decimal.NewFromInt(10),
		},
		RealStockUnits:    decimal.NewFromInt(40),
		Consumed:          decimal.NewFromInt(2),
		SupplyMovements:   []supplydomain.SupplyMovement{{IsEntry: true, Quantity: decimal.NewFromInt(40)}},
		HasRealStockCount: true,
		Base: shareddomain.Base{
			UpdatedAt: updatedAt,
		},
	})

	payload, err := json.Marshal(item)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	assert.NoError(t, err)

	assert.Equal(t, "2026-04-01T12:00:00Z", result["updated_at"])
}

func TestFromDomain_MarshalJSON_OmitsUpdatedAtWhenZero(t *testing.T) {
	item := FromDomain(&stockdomain.Stock{
		Supply: &supplydomain.Supply{
			Name:         "Urea",
			CategoryName: "Fertilizantes",
			UnitID:       2,
			UnitName:     "Kg",
			Price:        decimal.NewFromInt(10),
		},
		RealStockUnits:    decimal.NewFromInt(40),
		Consumed:          decimal.NewFromInt(2),
		SupplyMovements:   []supplydomain.SupplyMovement{{IsEntry: true, Quantity: decimal.NewFromInt(40)}},
		HasRealStockCount: true,
	})

	payload, err := json.Marshal(item)
	assert.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(payload, &result)
	assert.NoError(t, err)

	_, exists := result["updated_at"]
	assert.False(t, exists)
}
