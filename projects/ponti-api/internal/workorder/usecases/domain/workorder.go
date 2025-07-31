package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Workorder dominio, incluye ClassTypeID
type Workorder struct {
	Number        string
	ProjectID     int64
	FieldID       int64
	LotID         int64
	CropID        int64
	LaborID       int64
	ClassTypeID   int64
	Contractor    string
	Observations  string
	Date          time.Time
	InvestorID    int64
	EffectiveArea decimal.Decimal
	Items         []WorkorderItem
}

// WorkorderItem ...
type WorkorderItem struct {
	SupplyID  int64
	TotalUsed decimal.Decimal
	FinalDose decimal.Decimal
}

// Filtro
type WorkorderFilter struct {
	ProjectID *int64
	FieldID   *int64
}
