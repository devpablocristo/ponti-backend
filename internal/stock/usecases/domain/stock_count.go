package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/shopspring/decimal"
)

// StockCount registra un conteo físico manual append-only para un supply.
type StockCount struct {
	ID           int64
	SupplyID     int64
	CountedUnits decimal.Decimal
	CountedAt    time.Time
	Note         string
	shareddomain.Base
}
