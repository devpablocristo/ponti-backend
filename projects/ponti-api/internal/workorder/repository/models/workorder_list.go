package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type WorkorderListElement struct {
	Number        string          `gorm:"column:number;primaryKey"`
	ProjectName   string          `gorm:"column:project_name"`
	FieldName     string          `gorm:"column:field_name"`
	LotName       string          `gorm:"column:lot_name"`
	Date          time.Time       `gorm:"column:date"`
	CropName      string          `gorm:"column:crop_name"`
	LaborName     string          `gorm:"column:labor_name"`
	ClassTypeName string          `gorm:"column:class_type_name"`
	Contractor    string          `gorm:"column:contractor"`
	SurfaceHa     decimal.Decimal `gorm:"column:surface_ha"`
	InputName     string          `gorm:"column:input_name"`
	Consumption   decimal.Decimal `gorm:"column:consumption"`
	Category      string          `gorm:"column:category"`
	Dose          decimal.Decimal `gorm:"column:dose"`
	CostPerHa     decimal.Decimal `gorm:"column:cost_per_ha"`
	UnitPrice     decimal.Decimal `gorm:"column:unit_price"`
	TotalCost     decimal.Decimal `gorm:"column:total_cost"`
}

// TableName overrides the default table name for GORM.
func (WorkorderListElement) TableName() string {
	// Replace with your actual table or view name
	return "workorder_list_view"
}
