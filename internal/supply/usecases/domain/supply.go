package domain

import (
	classdomain "github.com/alphacodinggroup/ponti-backend/internal/class-type/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Supply struct {
	ID           int64
	ProjectID    int64
	Name         string
	UnitID       int64
	UnitName     string
	Price        decimal.Decimal
	CategoryID   int64
	CategoryName string
	Type         classdomain.ClassType

	shareddomain.Base // Audit fields
}
