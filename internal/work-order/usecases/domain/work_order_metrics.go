// Package domain contiene modelos de dominio para work orders.
package domain

import "github.com/shopspring/decimal"

// WorkOrderMetrics agrega las 4 métricas específicas para work orders.
type WorkOrderMetrics struct {
	SurfaceHa   decimal.Decimal // Superficie ejecutada total
	Liters      decimal.Decimal // Consumo en litros total
	Kilograms   decimal.Decimal // Consumo en kilos total
	DirectCost  decimal.Decimal // Costo directo total (labor + insumos)
	OrdersCount int64
}
