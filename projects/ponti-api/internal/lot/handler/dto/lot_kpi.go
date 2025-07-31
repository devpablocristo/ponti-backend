// handler/dto/kpis.go

package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	"github.com/shopspring/decimal"
)

type LotKPIsResponse struct {
	SeededArea     decimal.Decimal `json:"seeded_area"`
	HarvestedArea  decimal.Decimal `json:"harvested_area"`
	YieldTnPerHa   decimal.Decimal `json:"yield_tn_per_ha"`
	CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
}

func FromDomainKPIs(k *domain.LotKPIs) *LotKPIsResponse {
	return &LotKPIsResponse{
		SeededArea:     k.SeededArea,
		HarvestedArea:  k.HarvestedArea,
		YieldTnPerHa:   k.YieldTnPerHa,
		CostPerHectare: k.CostPerHectare,
	}
}
