package dto

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"

// WorkOrderRequest para POST y PUT
type WorkOrderRequest struct {
	Number       string             `json:"number" binding:"required"`
	ProjectID    int64              `json:"project_id" binding:"required"`
	FieldID      int64              `json:"field_id" binding:"required"`
	LotID        int64              `json:"lot_id" binding:"required"`
	CropID       int64              `json:"crop_id" binding:"required"`
	LaborID      int64              `json:"labor_id" binding:"required"`
	Contractor   string             `json:"contractor"`
	Observations string             `json:"observations"`
	Items        []WorkOrderItemDTO `json:"items" binding:"required,dive"`
}

type WorkOrderItemDTO struct {
	SupplyID      int64   `json:"supply_id" binding:"required"`
	TotalUsed     float64 `json:"total_used" binding:"required"`
	EffectiveArea float64 `json:"effective_area" binding:"required"`
	FinalDose     float64 `json:"final_dose" binding:"required"`
}

func (r *WorkOrderRequest) ToDomain() *domain.WorkOrder {
	o := &domain.WorkOrder{
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
		o.Items = append(o.Items, domain.WorkOrderItem{
			SupplyID:      it.SupplyID,
			TotalUsed:     it.TotalUsed,
			EffectiveArea: it.EffectiveArea,
			FinalDose:     it.FinalDose,
		})
	}
	return o
}

// WorkOrderDetail para GET /workorders/:number
type WorkOrderDetail struct {
	Number       string             `json:"number"`
	ProjectID    int64              `json:"project_id"`
	FieldID      int64              `json:"field_id"`
	LotID        int64              `json:"lot_id"`
	CropID       int64              `json:"crop_id"`
	LaborID      int64              `json:"labor_id"`
	Contractor   string             `json:"contractor"`
	Observations string             `json:"observations"`
	Items        []WorkOrderItemDTO `json:"items"`
}

func FromDomain(o *domain.WorkOrder) *WorkOrderDetail {
	d := &WorkOrderDetail{
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
		d.Items = append(d.Items, WorkOrderItemDTO{
			SupplyID:      it.SupplyID,
			TotalUsed:     it.TotalUsed,
			EffectiveArea: it.EffectiveArea,
			FinalDose:     it.FinalDose,
		})
	}
	return d
}

// WorkOrderResponse para POST (creación/duplicado)
type WorkOrderResponse struct {
	Message string `json:"message"`
	Number  string `json:"number"`
}
