package stock

import (
	"testing"
	"time"

	models "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestStockSummaryRow_ToDomain_UsesSupplyIDAsOperationalID(t *testing.T) {
	now := time.Now().UTC()
	row := models.StockSummaryRow{
		SupplyID:          7,
		ProjectID:         55,
		SupplyName:        "Urea",
		ClassType:         "Fertilizantes",
		SupplyUnitID:      2,
		SupplyUnitName:    "Kg",
		SupplyUnitPrice:   decimal.NewFromInt(10),
		EntryStock:        decimal.NewFromInt(20),
		OutStock:          decimal.NewFromInt(3),
		Consumed:          decimal.NewFromInt(5),
		StockUnits:        decimal.NewFromInt(12),
		RealStockUnits:    decimal.NewFromInt(11),
		HasRealStockCount: true,
		LastCountAt:       &now,
	}

	stock := row.ToDomain()

	assert.Equal(t, int64(7), stock.ID)
	assert.Equal(t, int64(55), stock.ProjectID)
	assert.Equal(t, "Urea", stock.Supply.Name)
	assert.True(t, stock.EntryStock.Equal(decimal.NewFromInt(20)))
	assert.True(t, stock.OutStock.Equal(decimal.NewFromInt(3)))
	assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(5)))
	assert.True(t, stock.StockUnits.Equal(decimal.NewFromInt(12)))
	assert.True(t, stock.RealStockUnits.Equal(decimal.NewFromInt(11)))
	assert.True(t, stock.HasRealStockCount)
	assert.Equal(t, &now, stock.LastCountAt)
}
