package domain

import "github.com/shopspring/decimal"

// Workorder representa una orden de trabajo
type Workorder struct {
	Number        string
	ProjectID     int64
	FieldID       int64
	LotID         int64
	CropID        int64
	LaborID       int64
	Contractor    string
	Observations  string
	Date          string
	InvestorID    int64
	EffectiveArea decimal.Decimal
	Items         []WorkorderItem
}

// WorkorderItem representa un insumo dentro de la orden
type WorkorderItem struct {
	SupplyID  int64
	TotalUsed decimal.Decimal
	FinalDose decimal.Decimal
}

// Filtro para listar workorders
type WorkorderFilter struct {
	ProjectID *int64
	FieldID   *int64
}
