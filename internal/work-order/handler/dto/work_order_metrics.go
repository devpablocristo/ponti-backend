// Package dto define los DTOs HTTP para work orders.
package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"

	domain "github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type WorkOrderMetrics struct {
	SurfaceHa  decimal.Decimal `json:"surface_ha"`  // Superficie ejecutada total
	Liters     decimal.Decimal `json:"liters"`      // Consumo en litros total
	Kilograms  decimal.Decimal `json:"kilograms"`   // Consumo en kilos total
	DirectCost decimal.Decimal `json:"direct_cost"` // Costo directo total (labor + insumos)
}

func (m WorkOrderMetrics) MarshalJSON() ([]byte, error) {
	aux := struct {
		SurfaceHa  string `json:"surface_ha"`
		Liters     string `json:"liters"`
		Kilograms  string `json:"kilograms"`
		DirectCost string `json:"direct_cost"`
	}{
		SurfaceHa:  m.SurfaceHa.Round(0).String(),  // Sin decimales
		Liters:     m.Liters.Round(0).String(),     // Sin decimales
		Kilograms:  m.Kilograms.Round(0).String(),  // Sin decimales
		DirectCost: m.DirectCost.Round(0).String(), // Sin decimales
	}
	return json.Marshal(aux)
}

func FromDomainMetrics(d *domain.WorkOrderMetrics) WorkOrderMetrics {
	return WorkOrderMetrics{
		SurfaceHa:  d.SurfaceHa,
		Liters:     d.Liters,
		Kilograms:  d.Kilograms,
		DirectCost: d.DirectCost,
	}
}
