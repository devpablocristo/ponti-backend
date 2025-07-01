// handler/dto/kpis.go

package dto

import domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"

type LotKPIsResponse struct {
	SeededArea     float64 `json:"seeded_area"`
	HarvestedArea  float64 `json:"harvested_area"`
	YieldTnPerHa   float64 `json:"yield_tn_per_ha"`
	CostPerHectare float64 `json:"cost_per_hectare"`
}

func FromDomainKPIs(k *domain.LotKPIs) *LotKPIsResponse {
	return &LotKPIsResponse{
		SeededArea:     k.SeededArea,
		HarvestedArea:  k.HarvestedArea,
		YieldTnPerHa:   k.YieldTnPerHa,
		CostPerHectare: k.CostPerHectare,
	}
}
