package models

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

type StockSummaryRow struct {
	SupplyID          int64           `gorm:"column:supply_id"`
	ProjectID         int64           `gorm:"column:project_id"`
	SupplyName        string          `gorm:"column:supply_name"`
	ClassType         string          `gorm:"column:class_type"`
	SupplyUnitID      int64           `gorm:"column:supply_unit_id"`
	SupplyUnitName    string          `gorm:"column:supply_unit_name"`
	SupplyUnitPrice   decimal.Decimal `gorm:"column:supply_unit_price"`
	EntryStock        decimal.Decimal `gorm:"column:entry_stock"`
	OutStock          decimal.Decimal `gorm:"column:out_stock"`
	Consumed          decimal.Decimal `gorm:"column:consumed"`
	StockUnits        decimal.Decimal `gorm:"column:stock_units"`
	RealStockUnits    decimal.Decimal `gorm:"column:real_stock_units"`
	HasRealStockCount bool            `gorm:"column:has_real_stock_count"`
	LastCountAt       *time.Time      `gorm:"column:last_count_at"`
}

func (r StockSummaryRow) ToDomain() *stockdomain.Stock {
	return &stockdomain.Stock{
		ID:        r.SupplyID,
		ProjectID: r.ProjectID,
		Supply: &supplydomain.Supply{
			ID:           r.SupplyID,
			Name:         r.SupplyName,
			UnitID:       r.SupplyUnitID,
			UnitName:     r.SupplyUnitName,
			Price:        r.SupplyUnitPrice,
			CategoryName: r.ClassType,
		},
		EntryStock:        r.EntryStock,
		OutStock:          r.OutStock,
		Consumed:          r.Consumed,
		StockUnits:        r.StockUnits,
		RealStockUnits:    r.RealStockUnits,
		HasRealStockCount: r.HasRealStockCount,
		LastCountAt:       r.LastCountAt,
		Base:              shareddomain.Base{},
	}
}
