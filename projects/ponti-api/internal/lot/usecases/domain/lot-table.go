package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type LotTable struct {
	ID             int64
	ProjectID      int64
	ProjectName    string
	FieldName      string
	LotName        string
	PreviousCrop   string
	PreviousCropID int64
	CurrentCrop    string
	CurrentCropID  int64
	Variety        string
	SowedArea      float64
	Season         string
	Tons           int
	Dates          []LotDates
	UpdatedAt      *time.Time
	AdminCost      decimal.Decimal
}

type LotDates struct {
	SowingDate  *time.Time
	HarvestDate *time.Time
	Sequence    int
}
