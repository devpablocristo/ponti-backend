package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type LaborListItem struct {
	WorkOrderID            int64
	WorkOrderNumber        string
	Date                   time.Time
	ProjectName            string
	FieldName              string
	LotId                  int64
	LotName                string
	CropName               string
	LaborName              string
	Contractor             string
	SurfaceHa              decimal.Decimal
	CostHa                 decimal.Decimal
	CategoryName           string
	InvestorID             int64
	InvestorName           string
	USDAvgValue            decimal.Decimal
	NetTotal               decimal.Decimal
	TotalIVA               decimal.Decimal
	USDCostHa              decimal.Decimal
	USDNetTotal            decimal.Decimal

	InvoiceID      int64
	InvoiceNumber  string
	InvoiceCompany string
	InvoiceDate    *time.Time
	InvoiceStatus  string
}

type LaborRawItem struct {
	WorkOrderID     int64
	WorkOrderNumber string
	Date            time.Time
	ProjectName     string
	FieldName       string
	CropName        string
	LaborName       string
	Contractor      string
	SurfaceHa       decimal.Decimal
	CostHa          decimal.Decimal
	CategoryName    string
	InvestorID      int64
	InvestorName    string
	USDAvgValue     decimal.Decimal

	// Campos calculados de la vista fix_labors_list
	NetTotal    decimal.Decimal
	TotalIVA    decimal.Decimal
	USDCostHa   decimal.Decimal
	USDNetTotal decimal.Decimal

	InvoiceID      int64
	InvoiceNumber  string
	InvoiceCompany string
	InvoiceDate    *time.Time
	InvoiceStatus  string
}
