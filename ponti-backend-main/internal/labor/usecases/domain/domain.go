package domain

import (
	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	"github.com/shopspring/decimal"
)

type Labor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          decimal.Decimal
	IsPartialPrice bool
	ProjectId      int64
	CategoryId     int64
	shareddomain.Base
}

type ListedLabor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          decimal.Decimal
	IsPartialPrice bool
	ProjectId      int64
	CategoryId     int64
	CategoryName   string

	shareddomain.Base
}
