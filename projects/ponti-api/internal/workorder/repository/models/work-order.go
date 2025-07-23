package models

import (
	"gorm.io/gorm"
)

// WorkOrder tabla principal
type WorkOrder struct {
	Number       string          `gorm:"primaryKey;column:number"`
	ProjectID    int64           `gorm:"not null"`
	FieldID      int64           `gorm:"not null"`
	LotID        int64           `gorm:"not null"`
	CropID       int64           `gorm:"not null"`
	LaborID      int64           `gorm:"not null"`
	Contractor   string          `gorm:"size:100"`
	Observations string          `gorm:"size:1000"`
	Items        []WorkOrderItem `gorm:"foreignKey:WorkOrderNumber;references:Number"`
	gorm.Model
}

// WorkOrderItem detalla los insumos
type WorkOrderItem struct {
	ID              int64   `gorm:"primaryKey;autoIncrement"`
	WorkOrderNumber string  `gorm:"column:order_number;index"`
	SupplyID        int64   `gorm:"not null"`
	TotalUsed       float64 `gorm:"not null"`
	EffectiveArea   float64 `gorm:"not null"`
	FinalDose       float64 `gorm:"not null"`
}
