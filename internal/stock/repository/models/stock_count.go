package models

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
)

type StockCount struct {
	ID           int64           `gorm:"primaryKey;autoIncrement;column:id"`
	SupplyID     int64           `gorm:"not null;index;column:supply_id"`
	CountedUnits decimal.Decimal `gorm:"not null;column:counted_units"`
	CountedAt    time.Time       `gorm:"not null;column:counted_at"`
	Note         string          `gorm:"column:note"`
	sharedmodels.Base
}

func (StockCount) TableName() string { return "supply_stock_counts" }

func StockCountFromDomain(d *stockdomain.StockCount) *StockCount {
	if d == nil {
		return nil
	}
	return &StockCount{
		ID:           d.ID,
		SupplyID:     d.SupplyID,
		CountedUnits: d.CountedUnits,
		CountedAt:    d.CountedAt,
		Note:         d.Note,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}

func (m *StockCount) ToDomain() *stockdomain.StockCount {
	if m == nil {
		return nil
	}
	return &stockdomain.StockCount{
		ID:           m.ID,
		SupplyID:     m.SupplyID,
		CountedUnits: m.CountedUnits,
		CountedAt:    m.CountedAt,
		Note:         m.Note,
		Base: shareddomain.Base{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			UpdatedBy: m.UpdatedBy,
		},
	}
}
