package domain

import (
	"testing"

	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetStockDifferencePtr_ReturnsNilWhenNoRealCount(t *testing.T) {
	stock := &Stock{
		Supply: &supplydomain.Supply{
			Name:     "2-4D",
			UnitName: "Lt",
			Price:    decimal.NewFromFloat(4.6),
		},
		Consumed:          decimal.NewFromInt(100),
		RealStockUnits:    decimal.Zero,
		HasRealStockCount: false,
	}

	assert.Equal(t, decimal.NewFromInt(-100), stock.GetStockUnits())
	assert.Nil(t, stock.GetStockDifferencePtr())
}

func TestGetStockDifferencePtr_ReturnsDifferenceWhenRealCountExists(t *testing.T) {
	stock := &Stock{
		Supply: &supplydomain.Supply{
			Name:     "Urea",
			UnitName: "Kg",
			Price:    decimal.NewFromInt(10),
		},
		SupplyMovements:   []supplydomain.SupplyMovement{{IsEntry: true, Quantity: decimal.NewFromInt(40)}},
		Consumed:          decimal.NewFromInt(2),
		RealStockUnits:    decimal.NewFromInt(40),
		HasRealStockCount: true,
	}

	diff := stock.GetStockDifferencePtr()
	if assert.NotNil(t, diff) {
		assert.True(t, diff.Equal(decimal.NewFromInt(2)))
	}
}
