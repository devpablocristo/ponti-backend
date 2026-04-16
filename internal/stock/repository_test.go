package stock

import (
	"testing"

	investormodels "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	models "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	supplymodels "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMapStockModelsToDomain_PreservesRowsPerInvestor(t *testing.T) {
	stockModels := []models.Stock{
		{
			ID:         1,
			SupplyID:   7,
			InvestorID: 10,
			Supply:     supplymodels.Supply{Name: "Urea"},
			Investor:   investormodels.Investor{Name: "Inv A"},
		},
		{
			ID:         2,
			SupplyID:   7,
			InvestorID: 11,
			Supply:     supplymodels.Supply{Name: "Urea"},
			Investor:   investormodels.Investor{Name: "Inv B"},
		},
	}

	stocks := mapStockModelsToDomain(stockModels, map[int64]decimal.Decimal{
		7: decimal.NewFromInt(5),
	})

	if assert.Len(t, stocks, 2) {
		assert.Equal(t, int64(1), stocks[0].ID)
		assert.Equal(t, "Inv A", stocks[0].Investor.Name)
		assert.Equal(t, int64(2), stocks[1].ID)
		assert.Equal(t, "Inv B", stocks[1].Investor.Name)
	}

	for _, stock := range stocks {
		assert.True(t, stock.Consumed.Equal(decimal.NewFromInt(5)))
	}
}
