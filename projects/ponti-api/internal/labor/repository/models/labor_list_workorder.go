package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type LaborRawItem struct {
	WorkorderNumber string          `gorm:"column:workorder_number"`
	Date            time.Time       `gorm:"column:date"`
	ProjectName     string          `gorm:"column:project_name"`
	FieldName       string          `gorm:"column:field_name"`
	CropName        string          `gorm:"column:crop_name"`
	LaborName       string          `gorm:"column:labor_name"`
	Contractor      string          `gorm:"column:contractor"`
	SurfaceHa       decimal.Decimal `gorm:"column:effective_area"`
	CostHa          decimal.Decimal `gorm:"column:price"`
	CategoryName    string          `gorm:"column:contractor_name"`
	InvestorName    string          `gorm:"column:investor_name"`
	USDAvgValue     decimal.Decimal `gorm:"column:usd_avg_value"`
	InvoiceNumber   string          `gorm:"column:invoice_number"`
	InvoiceCompany  string          `gorm:"column:invoice_company"`
	InvoiceDate     time.Time       `gorm:"column:invoice_date"`
	InvoiceStatus   string          `gorm:"column:invoice_status"`
}
