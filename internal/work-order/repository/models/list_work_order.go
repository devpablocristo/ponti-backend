// Package models contiene modelos de persistencia para work orders.
package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkOrderListElement mapea directamente la vista SQL.
type WorkOrderListElement struct {
	ID                int64           `gorm:"column:id" json:"id"`
	Number            string          `gorm:"column:number" json:"number"`
	ProjectName       string          `gorm:"column:project_name" json:"project_name"`
	FieldName         string          `gorm:"column:field_name" json:"field_name"`
	LotName           string          `gorm:"column:lot_name" json:"lot_name"`
	Date              time.Time       `gorm:"column:date" json:"date"`
	CropName          string          `gorm:"column:crop_name" json:"crop_name"`
	LaborName         string          `gorm:"column:labor_name" json:"labor_name"`
	LaborCategoryName string          `gorm:"column:labor_category_name" json:"labor_category_name"`
	TypeName          string          `gorm:"column:type_name" json:"type_name"`
	Contractor        string          `gorm:"column:contractor" json:"contractor"`
	SurfaceHa         decimal.Decimal `gorm:"column:surface_ha" json:"surface_ha"`
	SupplyName        string          `gorm:"column:supply_name" json:"supply_name"`
	Consumption       decimal.Decimal `gorm:"column:consumption" json:"consumption"`
	CategoryName      string          `gorm:"column:category_name" json:"category_name"`
	Dose              decimal.Decimal `gorm:"column:dose_per_ha" json:"dose"`
	CostPerHa         decimal.Decimal `gorm:"column:supply_cost_per_ha" json:"cost_per_ha"`
	UnitPrice         decimal.Decimal `gorm:"column:unit_price" json:"unit_price"`
	TotalCost         decimal.Decimal `gorm:"column:supply_total_cost" json:"total_cost"`
	IsDigital         bool            `gorm:"column:is_digital" json:"is_digital"`
	Status            string          `gorm:"column:status" json:"status"`
}

// TableName apunta a la vista.
func (WorkOrderListElement) TableName() string {
	return "v4_report.workorder_list"
}
