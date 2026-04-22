package domain

import (
	"testing"
	"time"

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
		StockUnits:        decimal.NewFromInt(100),
		RealStockUnits:    decimal.Zero,
		HasRealStockCount: false,
	}

	assert.True(t, stock.GetStockUnits().Equal(decimal.NewFromInt(100)))
	assert.Nil(t, stock.GetStockDifferencePtr())
}

func TestGetStockDifferencePtr_ReturnsDifferenceWhenRealCountExists(t *testing.T) {
	now := time.Now().UTC()
	stock := &Stock{
		Supply: &supplydomain.Supply{
			Name:     "Urea",
			UnitName: "Kg",
			Price:    decimal.NewFromInt(10),
		},
		StockUnits:        decimal.NewFromInt(38),
		RealStockUnits:    decimal.NewFromInt(40),
		HasRealStockCount: true,
		LastCountAt:       &now,
	}

	diff := stock.GetStockDifferencePtr()
	if assert.NotNil(t, diff) {
		assert.True(t, diff.Equal(decimal.NewFromInt(2)))
	}
}
