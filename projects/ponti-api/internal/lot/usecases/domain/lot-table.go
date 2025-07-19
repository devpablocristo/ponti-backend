package domain

import "time"

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
	Dates          []LotDates
	UpdatedAt      *time.Time
	CostPerHectare float64
}

type LotDates struct {
	SowingDate  *time.Time
	HarvestDate *time.Time
	Sequence    int
}
