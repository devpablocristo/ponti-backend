// Package domain contiene modelos de dominio para work orders.
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkOrderListElement para la lista.
type WorkOrderListElement struct {
	ID                int64
	Number            string
	ProjectName       string
	FieldName         string
	LotName           string
	Date              time.Time
	CropName          string
	LaborName         string
	LaborCategoryName string
	TypeName          string
	Contractor        string
	SurfaceHa         decimal.Decimal
	SupplyName        string
	Consumption       decimal.Decimal
	CategoryName      string
	Dose              decimal.Decimal
	CostPerHa         decimal.Decimal
	UnitPrice         decimal.Decimal
	TotalCost         decimal.Decimal
	IsDigital         bool
	Status            string
}

// - CropName: Estos filtros muestran una lista desplegable con todas las opciones de cultivos y un checkbox a la par para activar los cultivos que se quieran ver. Lo mismo es para cada una de las columnas.
//
// - Configurar Columnas: Esto debería mostrar una lista de los titulos de columnas disponibles con un checkbox para ir chequeando las que queremos ver activas. Por default, todas deben estar activadas.
