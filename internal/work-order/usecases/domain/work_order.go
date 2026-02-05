// Package domain contiene modelos de dominio para work orders.
package domain

import (
	"time"

	"github.com/shopspring/decimal"

	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
)

// WorkOrder representa una orden de trabajo.
type WorkOrder struct {
	ID            int64
	Number        string
	ProjectID     int64
	FieldID       int64
	LotID         int64
	CropID        int64
	LaborID       int64
	Contractor    string
	Observations  string
	Date          time.Time
	InvestorID    int64
	EffectiveArea decimal.Decimal
	Items         []WorkOrderItem

	Base shareddomain.Base
}

// WorkOrderItem representa un item de la orden de trabajo.
type WorkOrderItem struct {
	SupplyID  int64
	TotalUsed decimal.Decimal
	FinalDose decimal.Decimal
}

// WorkOrderFilter representa filtros para listar órdenes de trabajo.
type WorkOrderFilter struct {
	ProjectID  *int64
	FieldID    *int64
	CustomerID *int64
	CampaignID *int64
}
