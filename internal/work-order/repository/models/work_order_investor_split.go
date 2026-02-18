package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// WorkOrderInvestorSplit almacena porcentajes por inversor dentro de una OT.
// Se usa para "Dividir aporte" sin duplicar workorders.
type WorkOrderInvestorSplit struct {
	ID          int64           `gorm:"primaryKey;autoIncrement;column:id"`
	WorkOrderID int64           `gorm:"column:workorder_id;index;not null"`
	InvestorID  int64           `gorm:"column:investor_id;index;not null"`
	Percentage  decimal.Decimal `gorm:"column:percentage;not null"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"`
}

func (WorkOrderInvestorSplit) TableName() string { return "workorder_investor_splits" }
