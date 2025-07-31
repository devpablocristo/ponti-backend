package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkorderListElement representa la vista a nivel de base para listar órdenes de trabajo.
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
	SupplyName    string          `gorm:"column:supply_name"`
	Consumption   decimal.Decimal `gorm:"column:consumption"`
	CategoryName  string          `gorm:"column:category_name"`
	Dose          decimal.Decimal `gorm:"column:dose"`
	CostPerHa     decimal.Decimal `gorm:"column:cost_per_ha"`
	UnitPrice     decimal.Decimal `gorm:"column:unit_price"`
	TotalCost     decimal.Decimal `gorm:"column:total_cost"`
}

// TableName especifica la tabla o vista en la base de datos.
func (WorkorderListElement) TableName() string {
	return "workorder_list_view"
}
