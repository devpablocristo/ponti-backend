package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Supply struct {
	ID         int64
	ProjectID  int64
	Name       string
	UnitID     int64
	Price      decimal.Decimal
	CategoryID int64
	TypeID     int64

	shareddomain.Base // Audit fields
}
