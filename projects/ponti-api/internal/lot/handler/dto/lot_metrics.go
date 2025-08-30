package dto

import (
	"encoding/json"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	"github.com/shopspring/decimal"
)

type LotMetrics struct {
	SeededArea     decimal.Decimal `json:"seeded_area"`
	HarvestedArea  decimal.Decimal `json:"harvested_area"`
	YieldTnPerHa   decimal.Decimal `json:"yield_tn_per_ha"`
	CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
}

// Redondeo homogéneo (3 decimales) como en workorders.
func (m LotMetrics) MarshalJSON() ([]byte, error) {
	aux := struct {
		SeededArea     string `json:"seeded_area"`
		HarvestedArea  string `json:"harvested_area"`
		YieldTnPerHa   string `json:"yield_tn_per_ha"`
		CostPerHectare string `json:"cost_per_hectare"`
	}{
		SeededArea:     m.SeededArea.Round(3).String(),
		HarvestedArea:  m.HarvestedArea.Round(3).String(),
		YieldTnPerHa:   m.YieldTnPerHa.Round(3).String(),
		CostPerHectare: m.CostPerHectare.Round(3).String(),
	}
	return json.Marshal(aux)
}

func FromDomainMetrics(d *domain.LotMetrics) LotMetrics {
	return LotMetrics{
		SeededArea:     d.SeededArea,
		HarvestedArea:  d.HarvestedArea,
		YieldTnPerHa:   d.YieldTnPerHa,
		CostPerHectare: d.CostPerHectare,
	}
}
