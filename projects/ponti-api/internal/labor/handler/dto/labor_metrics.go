package dto

import (
	"encoding/json"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/shopspring/decimal"
)

type LaborMetrics struct {
	SurfaceHa    decimal.Decimal `json:"surface_ha"`
	NetTotalCost decimal.Decimal `json:"net_total_cost"`
	AvgCostPerHa decimal.Decimal `json:"avg_cost_per_ha"`
}

func (m LaborMetrics) MarshalJSON() ([]byte, error) {
	aux := struct {
		SurfaceHa    string `json:"surface_ha"`
		NetTotalCost string `json:"net_total_cost"`
		AvgCostPerHa string `json:"avg_cost_per_ha"`
	}{
		SurfaceHa:    m.SurfaceHa.Round(3).String(),
		NetTotalCost: m.NetTotalCost.Round(3).String(),
		AvgCostPerHa: m.AvgCostPerHa.Round(3).String(),
	}
	return json.Marshal(aux)
}

func FromDomainMetrics(d *domain.LaborMetrics) LaborMetrics {
	return LaborMetrics{
		SurfaceHa:    d.SurfaceHa,
		NetTotalCost: d.NetTotalCost,
		AvgCostPerHa: d.AvgCostPerHa,
	}
}
