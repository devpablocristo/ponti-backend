package dto

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"

// WorkorderRequest para POST y PUT
type WorkorderRequest struct {
	Number       string             `json:"number" binding:"required"`
	ProjectID    int64              `json:"project_id" binding:"required"`
	FieldID      int64              `json:"field_id" binding:"required"`
	LotID        int64              `json:"lot_id" binding:"required"`
	CropID       int64              `json:"crop_id" binding:"required"`
	LaborID      int64              `json:"labor_id" binding:"required"`
	Contractor   string             `json:"contractor"`
	Observations string             `json:"observations"`
	Items        []WorkorderItemDTO `json:"items" binding:"required,dive"`
}

type WorkorderItemDTO struct {
	SupplyID      int64   `json:"supply_id" binding:"required"`
	TotalUsed     float64 `json:"total_used" binding:"required"`
	EffectiveArea float64 `json:"effective_area" binding:"required"`
	FinalDose     float64 `json:"final_dose" binding:"required"`
}

func (r *WorkorderRequest) ToDomain() *domain.Workorder {
	o := &domain.Workorder{
		Number:       r.Number,
		ProjectID:    r.ProjectID,
		FieldID:      r.FieldID,
		LotID:        r.LotID,
		CropID:       r.CropID,
		LaborID:      r.LaborID,
		Contractor:   r.Contractor,
		Observations: r.Observations,
	}
	for _, it := range r.Items {
		o.Items = append(o.Items, domain.WorkorderItem{
			SupplyID:      it.SupplyID,
			TotalUsed:     it.TotalUsed,
			EffectiveArea: it.EffectiveArea,
			FinalDose:     it.FinalDose,
		})
	}
	return o
}

// WorkorderDetail para GET /workorders/:number
type WorkorderDetail struct {
	Number       string             `json:"number"`
	ProjectID    int64              `json:"project_id"`
	FieldID      int64              `json:"field_id"`
	LotID        int64              `json:"lot_id"`
	CropID       int64              `json:"crop_id"`
	LaborID      int64              `json:"labor_id"`
	Contractor   string             `json:"contractor"`
	Observations string             `json:"observations"`
	Items        []WorkorderItemDTO `json:"items"`
}

func FromDomain(o *domain.Workorder) *WorkorderDetail {
	d := &WorkorderDetail{
		Number:       o.Number,
		ProjectID:    o.ProjectID,
		FieldID:      o.FieldID,
		LotID:        o.LotID,
		CropID:       o.CropID,
		LaborID:      o.LaborID,
		Contractor:   o.Contractor,
		Observations: o.Observations,
	}
	for _, it := range o.Items {
		d.Items = append(d.Items, WorkorderItemDTO{
			SupplyID:      it.SupplyID,
			TotalUsed:     it.TotalUsed,
			EffectiveArea: it.EffectiveArea,
			FinalDose:     it.FinalDose,
		})
	}
	return d
}

// WorkorderResponse para POST (creación/duplicado)
type WorkorderResponse struct {
	Message string `json:"message"`
	Number  string `json:"number"`
}
