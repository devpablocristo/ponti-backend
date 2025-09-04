package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type LaborRawItem struct {
	WorkorderID     int64           `gorm:"column:workorder_id"`
	WorkorderNumber string          `gorm:"column:workorder_number"`
	Date            time.Time       `gorm:"column:date"`
	ProjectName     string          `gorm:"column:project_name"`
	FieldName       string          `gorm:"column:field_name"`
	CropName        string          `gorm:"column:crop_name"`
	LaborName       string          `gorm:"column:labor_name"`
	Contractor      string          `gorm:"column:contractor"`
	SurfaceHa       decimal.Decimal `gorm:"column:surface_ha"`
	CostHa          decimal.Decimal `gorm:"column:cost_ha"`
	CategoryName    string          `gorm:"column:category_name"`
	InvestorName    string          `gorm:"column:investor_name"`
	USDAvgValue     decimal.Decimal `gorm:"column:usd_avg_value"`

	// Campos calculados de la vista fix_labors_list
	NetTotal    decimal.Decimal `gorm:"column:net_total"`
	TotalIVA    decimal.Decimal `gorm:"column:total_iva"`
	USDCostHa   decimal.Decimal `gorm:"column:usd_cost_ha"`
	USDNetTotal decimal.Decimal `gorm:"column:usd_net_total"`

	InvoiceID      int64     `gorm:"column:invoice_id"`
	InvoiceNumber  string    `gorm:"column:invoice_number"`
	InvoiceCompany string    `gorm:"column:invoice_company"`
	InvoiceDate    time.Time `gorm:"column:invoice_date"`
	InvoiceStatus  string    `gorm:"column:invoice_status"`
}
