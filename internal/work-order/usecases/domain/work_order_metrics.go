// Package domain contiene modelos de dominio para work orders.
package domain

import "github.com/shopspring/decimal"

// WorkOrderMetrics agrega las 4 métricas específicas para work orders.
type WorkOrderMetrics struct {
	SurfaceHa  decimal.Decimal `json:"surface_ha"`  // Superficie ejecutada total
	Liters     decimal.Decimal `json:"liters"`      // Consumo en litros total
	Kilograms  decimal.Decimal `json:"kilograms"`   // Consumo en kilos total
	DirectCost decimal.Decimal `json:"direct_cost"` // Costo directo total (labor + insumos)
}
