package dto

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"

// WorkOrder body
type WorkOrder struct {
	Number       string          `json:"number" binding:"required"`
	ProjectID    int64           `json:"project_id" binding:"required"`
	FieldID      int64           `json:"field_id" binding:"required"`
	LotID        int64           `json:"lot_id" binding:"required"`
	CropID       int64           `json:"crop_id" binding:"required"`
	LaborID      int64           `json:"labor_id" binding:"required"`
	Contractor   string          `json:"contractor"`
	Observations string          `json:"observations"`
	Items        []WorkOrderItem `json:"items" binding:"required,dive"`
}

// WorkOrderItem representa un insumo en la petición
type WorkOrderItem struct {
	SupplyID      int64   `json:"supply_id" binding:"required"`
	TotalUsed     float64 `json:"total_used" binding:"required"`
	EffectiveArea float64 `json:"effective_area" binding:"required"`
	FinalDose     float64 `json:"final_dose" binding:"required"`
}

func (c *WorkOrder) ToDomain() *domain.WorkOrder {
	o := &domain.WorkOrder{
		Number:       c.Number,
		ProjectID:    c.ProjectID,
		FieldID:      c.FieldID,
		LotID:        c.LotID,
		CropID:       c.CropID,
		LaborID:      c.LaborID,
		Contractor:   c.Contractor,
		Observations: c.Observations,
	}
	for _, item := range c.Items {
		o.Items = append(o.Items, domain.WorkOrderItem{
			SupplyID:      item.SupplyID,
			TotalUsed:     item.TotalUsed,
			EffectiveArea: item.EffectiveArea,
			FinalDose:     item.FinalDose,
		})
	}
	return o
}

// WorkOrder respuesta tras creación
type WorkOrderResponse struct {
	Message string `json:"message"`
	Number  string `json:"number"`
}
