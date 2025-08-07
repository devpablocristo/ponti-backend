package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type LaborListItem struct {
	WorkorderNumber string          //numero de orden
	Date            time.Time       // fecha de la orden
	ProjectName     string          // nombre del proyecto
	FieldName       string          // nombre del campo
	CropName        string          // nombre del cultivo
	LaborName       string          // nombre de la labor
	Contractor      string          // nombre del contratista
	SurfaceHa       decimal.Decimal // superficie
	CostHa          decimal.Decimal // costo por ha
	CategoryName    string          // rubro
	InvestorName    string          // nombre del inversor
	NetTotal        decimal.Decimal // total neto
	TotalIVA        decimal.Decimal // total IVA
}

// este tipo representa a los datos crudos extraidos antes para ser calculados luego
type LaborRawItem struct {
	WorkorderNumber string
	Date            time.Time
	ProjectName     string
	FieldName       string
	CropName        string
	LaborName       string
	Contractor      string
	SurfaceHa       decimal.Decimal
	CostHa          decimal.Decimal
	CategoryName    string
	InvestorName    string
}
