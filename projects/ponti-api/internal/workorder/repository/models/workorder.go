package models

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
	"github.com/shopspring/decimal"
)

type Workorder struct {
	Number        string          `gorm:"primaryKey;column:number;uniqueIndex"`
	ProjectID     int64           `gorm:"not null"`
	FieldID       int64           `gorm:"not null"`
	LotID         int64           `gorm:"not null"`
	CropID        int64           `gorm:"not null"`
	LaborID       int64           `gorm:"not null"`
	Contractor    string          `gorm:"size:100"`
	Observations  string          `gorm:"size:1000"`
	Date          string          `gorm:"type:date;not null"`
	InvestorID    int64           `gorm:"not null"`
	EffectiveArea decimal.Decimal `gorm:"not null"`
	Items         []WorkorderItem `gorm:"foreignKey:WorkorderNumber;references:Number"`
}

type WorkorderItem struct {
	ID              int64           `gorm:"primaryKey;autoIncrement"`
	WorkorderNumber string          `gorm:"column:order_number;index"`
	SupplyID        int64           `gorm:"not null"`
	TotalUsed       decimal.Decimal `gorm:"not null"`
	FinalDose       decimal.Decimal `gorm:"not null"`
}

// FromDomain mapea domain.Workorder a models.Workorder
func FromDomain(o *domain.Workorder) *Workorder {
	items := make([]WorkorderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = WorkorderItem{
			WorkorderNumber: o.Number,
			SupplyID:        it.SupplyID,
			TotalUsed:       it.TotalUsed,
			FinalDose:       it.FinalDose,
		}
	}
	return &Workorder{
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
		Items:         items,
	}
}

// ToDomain mapea models.Workorder a domain.Workorder
func (m *Workorder) ToDomain() *domain.Workorder {
	items := make([]domain.WorkorderItem, len(m.Items))
	for i, it := range m.Items {
		items[i] = domain.WorkorderItem{
			SupplyID:  it.SupplyID,
			TotalUsed: it.TotalUsed,
			FinalDose: it.FinalDose,
		}
	}
	return &domain.Workorder{
		Number:        m.Number,
		ProjectID:     m.ProjectID,
		FieldID:       m.FieldID,
		LotID:         m.LotID,
		CropID:        m.CropID,
		LaborID:       m.LaborID,
		Contractor:    m.Contractor,
		Observations:  m.Observations,
		Date:          m.Date,
		InvestorID:    m.InvestorID,
		EffectiveArea: m.EffectiveArea,
		Items:         items,
	}
}
