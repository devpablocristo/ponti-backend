package models

import (
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
)

type StockUpdateRealUnits struct {
	RealStockUnits decimal.Decimal `gorm:"column:real_stock_units"`
	UpdatedBy      string          `gorm:"column:updated_by"`
}

func StockUpdateRealUnitsFromDomain(d *domain.Stock) *StockUpdateRealUnits {
	return &StockUpdateRealUnits{
		RealStockUnits: d.RealStockUnits,
		UpdatedBy:      *d.UpdatedBy,
	}
}
