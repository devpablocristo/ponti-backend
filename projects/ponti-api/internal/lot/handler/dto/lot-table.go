package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type LotTable struct {
	ID             int64           `json:"id"`
	ProjectID      int64           `json:"project_id"`
	PreviousCropID int64           `json:"previous_crop_id"`
	CurrentCropID  int64           `json:"current_crop_id"`
	ProjectName    string          `json:"project_name"`
	FieldName      string          `json:"field_name"`
	LotName        string          `json:"lot_name"`
	PreviousCrop   string          `json:"previous_crop"`
	CurrentCrop    string          `json:"current_crop"`
	Variety        string          `json:"variety"`
	SowedArea      float64         `json:"sowed_area"`
	Season         string          `json:"season"`
	Tons           int             `json:"tons"`
	Dates          []LotDates      `json:"dates"` // ISO 8601 o el formato que uses
	AdminCost      decimal.Decimal `json:"admin_cost"`
	UpdatedAt      *time.Time      `json:"updated_at,omitempty"`
}

type LotDates struct {
	SowingDate  string `json:"sowing_date"`
	HarvestDate string `json:"harvest_date"`
	Sequence    int    `json:"sequence"`
}

type LotTableResponse struct {
	Rows         []LotTable `json:"rows"`
	Total        int        `json:"total"` // total de registros sin paginar
	SumSowedArea float64    `json:"sum_sowed_area"`
	SumCost      float64    `json:"sum_cost"`
}
