package dto

import (
	"time"

	utils "github.com/alphacodinggroup/ponti-backend/pkg/utils"

	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type WorkorderItem struct {
	SupplyID  int64           `json:"supply_id" binding:"required"`
	TotalUsed decimal.Decimal `json:"total_used" binding:"required"`
	FinalDose decimal.Decimal `json:"final_dose" binding:"required"`
}

type Workorder struct {
	ID            int64           `json:"id"`
	Number        string          `json:"number" binding:"required"`
	ProjectID     int64           `json:"project_id" binding:"required"`
	FieldID       int64           `json:"field_id" binding:"required"`
	LotID         int64           `json:"lot_id" binding:"required"`
	CropID        int64           `json:"crop_id" binding:"required"`
	LaborID       int64           `json:"labor_id" binding:"required"`
	Contractor    string          `json:"contractor"`
	Observations  string          `json:"observations"`
	Date          utils.ISODate   `json:"date" binding:"required"`
	InvestorID    int64           `json:"investor_id" binding:"required"`
	EffectiveArea decimal.Decimal `json:"effective_area" binding:"required"`
	Items         []WorkorderItem `json:"items" binding:"required,dive"`
}

func (r *Workorder) ToDomain() *domain.Workorder {
	return &domain.Workorder{
		ID:            r.ID,
		Number:        r.Number,
		ProjectID:     r.ProjectID,
		FieldID:       r.FieldID,
		LotID:         r.LotID,
		CropID:        r.CropID,
		LaborID:       r.LaborID,
		Contractor:    r.Contractor,
		Observations:  r.Observations,
		Date:          time.Time(r.Date),
		InvestorID:    r.InvestorID,
		EffectiveArea: r.EffectiveArea,
		Items:         toDomainItems(r.Items),
	}
}

func toDomainItems(items []WorkorderItem) []domain.WorkorderItem {
	out := make([]domain.WorkorderItem, len(items))
	for i, it := range items {
		out[i] = domain.WorkorderItem{
			SupplyID:  it.SupplyID,
			TotalUsed: it.TotalUsed,
			FinalDose: it.FinalDose,
		}
	}
	return out
}

func FromDomain(o *domain.Workorder) *Workorder {
	items := make([]WorkorderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = WorkorderItem{
			SupplyID:  it.SupplyID,
			TotalUsed: it.TotalUsed,
			FinalDose: it.FinalDose,
		}
	}
	return &Workorder{
		ID:            o.ID,
		Number:        o.Number,
		ProjectID:     o.ProjectID,
		FieldID:       o.FieldID,
		LotID:         o.LotID,
		CropID:        o.CropID,
		LaborID:       o.LaborID,
		Contractor:    o.Contractor,
		Observations:  o.Observations,
		Date:          utils.ISODate(o.Date),
		InvestorID:    o.InvestorID,
		EffectiveArea: o.EffectiveArea,
		Items:         items,
	}
}

type WorkorderResponse struct {
	Message string `json:"message"`
	Number  string `json:"number"`
}
