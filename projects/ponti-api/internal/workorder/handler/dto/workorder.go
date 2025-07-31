package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
	"github.com/shopspring/decimal"
)

type WorkorderItem struct {
	SupplyID  int64           `json:"supply_id" binding:"required"`
	TotalUsed decimal.Decimal `json:"total_used" binding:"required"`
	FinalDose decimal.Decimal `json:"final_dose" binding:"required"`
}

type Workorder struct {
	Number        string          `json:"number" binding:"required"`
	ProjectID     int64           `json:"project_id" binding:"required"`
	FieldID       int64           `json:"field_id" binding:"required"`
	LotID         int64           `json:"lot_id" binding:"required"`
	CropID        int64           `json:"crop_id" binding:"required"`
	LaborID       int64           `json:"labor_id" binding:"required"`
	Contractor    string          `json:"contractor"`
	Observations  string          `json:"observations"`
	Date          string          `json:"date" binding:"required"`
	InvestorID    int64           `json:"investor_id" binding:"required"`
	EffectiveArea decimal.Decimal `json:"effective_area" binding:"required"`
	Items         []WorkorderItem `json:"items" binding:"required,dive"`
}

func (r *Workorder) ToDomain() *domain.Workorder {
	o := &domain.Workorder{
		Number:        r.Number,
		ProjectID:     r.ProjectID,
		FieldID:       r.FieldID,
		LotID:         r.LotID,
		CropID:        r.CropID,
		LaborID:       r.LaborID,
		Contractor:    r.Contractor,
		Observations:  r.Observations,
		Date:          r.Date,
		InvestorID:    r.InvestorID,
		EffectiveArea: r.EffectiveArea,
	}
	for _, it := range r.Items {
		o.Items = append(o.Items, domain.WorkorderItem{
			SupplyID:  it.SupplyID,
			TotalUsed: it.TotalUsed,
			FinalDose: it.FinalDose,
		})
	}
	return o
}

type WorkorderDetail struct {
	Number        string          `json:"number"`
	ProjectID     int64           `json:"project_id"`
	FieldID       int64           `json:"field_id"`
	LotID         int64           `json:"lot_id"`
	CropID        int64           `json:"crop_id"`
	LaborID       int64           `json:"labor_id"`
	Contractor    string          `json:"contractor"`
	Observations  string          `json:"observations"`
	Date          string          `json:"date"`
	InvestorID    int64           `json:"investor_id"`
	EffectiveArea decimal.Decimal `json:"effective_area"`
	Items         []WorkorderItem `json:"items"`
}

func FromDomain(o *domain.Workorder) *WorkorderDetail {
	d := &WorkorderDetail{
		Number:        o.Number,
		ProjectID:     o.ProjectID,
		FieldID:       o.FieldID,
		LotID:         o.LotID,
		CropID:        o.CropID,
		LaborID:       o.LaborID,
		Contractor:    o.Contractor,
		Observations:  o.Observations,
		Date:          o.Date,
		InvestorID:    o.InvestorID,
		EffectiveArea: o.EffectiveArea,
	}
	for _, it := range o.Items {
		d.Items = append(d.Items, WorkorderItem{
			SupplyID:  it.SupplyID,
			TotalUsed: it.TotalUsed,
			FinalDose: it.FinalDose,
		})
	}
	return d
}

type WorkorderResponse struct {
	Message string `json:"message"`
	Number  string `json:"number"`
}
