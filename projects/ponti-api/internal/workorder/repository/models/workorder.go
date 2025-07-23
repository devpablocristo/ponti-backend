package models

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"

// Workorder tabla principal usando Number como primary key,
// sin embebido de gorm.Model.
type Workorder struct {
	Number       string          `gorm:"primaryKey;column:number;uniqueIndex"`
	ProjectID    int64           `gorm:"not null"`
	FieldID      int64           `gorm:"not null"`
	LotID        int64           `gorm:"not null"`
	CropID       int64           `gorm:"not null"`
	LaborID      int64           `gorm:"not null"`
	Contractor   string          `gorm:"size:100"`
	Observations string          `gorm:"size:1000"`
	Items        []WorkorderItem `gorm:"foreignKey:WorkorderNumber;references:Number"`
}

// WorkorderItem detalla los insumos de la orden
type WorkorderItem struct {
	ID              int64   `gorm:"primaryKey;autoIncrement"`
	WorkorderNumber string  `gorm:"column:order_number;index"`
	SupplyID        int64   `gorm:"not null"`
	TotalUsed       float64 `gorm:"not null"`
	EffectiveArea   float64 `gorm:"not null"`
	FinalDose       float64 `gorm:"not null"`
}

// FromDomain mapea un dominio Workorder a su equivalente en models.Workorder
func FromDomain(o *domain.Workorder) *Workorder {
	items := make([]WorkorderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = WorkorderItem{
			WorkorderNumber: o.Number,
			SupplyID:        it.SupplyID,
			TotalUsed:       it.TotalUsed,
			EffectiveArea:   it.EffectiveArea,
			FinalDose:       it.FinalDose,
		}
	}
	return &Workorder{
		Number:       o.Number,
		ProjectID:    o.ProjectID,
		FieldID:      o.FieldID,
		LotID:        o.LotID,
		CropID:       o.CropID,
		LaborID:      o.LaborID,
		Contractor:   o.Contractor,
		Observations: o.Observations,
		Items:        items,
	}
}

// ToDomain mapea un models.Workorder (incluyendo sus items) a domain.Workorder
func (m *Workorder) ToDomain() *domain.Workorder {
	items := make([]domain.WorkorderItem, len(m.Items))
	for i, it := range m.Items {
		items[i] = domain.WorkorderItem{
			SupplyID:      it.SupplyID,
			TotalUsed:     it.TotalUsed,
			EffectiveArea: it.EffectiveArea,
			FinalDose:     it.FinalDose,
		}
	}
	return &domain.Workorder{
		Number:       m.Number,
		ProjectID:    m.ProjectID,
		FieldID:      m.FieldID,
		LotID:        m.LotID,
		CropID:       m.CropID,
		LaborID:      m.LaborID,
		Contractor:   m.Contractor,
		Observations: m.Observations,
		Items:        items,
	}
}
