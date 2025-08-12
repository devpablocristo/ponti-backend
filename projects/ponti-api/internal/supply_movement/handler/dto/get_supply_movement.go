package dto

import (
	"fmt"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"github.com/shopspring/decimal"
)

type GetEntrySupplyMovementsResponse struct {
	Summary summary `json:"summary"`
	EntrySupplyMovementsResponses []entrySupplyMovementsResponse `json:"entries"`
}

type summary struct{
	TotalKg  float64 `json:"total_kg"`
	TotalLt  float64 `json:"total_lt"`
	TotalUSD decimal.Decimal `json:"total_usd"`
}

type entrySupplyMovementsResponse struct{
	ID int64 `json:"id"`
	EntryType string `json:"entry_type"`
	ReferenceNumber string `json:"reference_number"`
	EntryDate time.Time `json:"entry_date"`
	InvestorName string `json:"investor_name"`
	SupplyName string `json:"supply_name"`
	Quantity string `json:"quantity"`
	Category string `json:"category"`
	Type string `json:"type"`
	ProviderName string `json:"provider_name"`
	PriceUSD decimal.Decimal `json:"price_usd"`
	TotalUSD decimal.Decimal `json:"total_usd"`

}

func entrySupplyMovementsResponseFromDomain(dsm *domain.SupplyMovement)  entrySupplyMovementsResponse{
	return entrySupplyMovementsResponse{
		ID: dsm.ID,
		EntryType: dsm.MovementType,
		ReferenceNumber: dsm.ReferenceNumber,
		EntryDate: *dsm.MovementDate,
		InvestorName: dsm.Investor.Name,
		SupplyName: dsm.Supply.Name,
		Quantity: fmt.Sprintf("%s %s", dsm.Quantity.String(), dsm.Supply.UnitName),
		Category: dsm.Supply.CategoryName,
		Type: dsm.Supply.Type.Name,
		PriceUSD: dsm.Supply.Price,
		TotalUSD: dsm.Supply.Price.Mul(dsm.Quantity),
	}
}

func NewGetEntrySupplyMovementsResponse(entriesDomain []*domain.SupplyMovement) GetEntrySupplyMovementsResponse {
	var totalKg float64
	var totalLt float64
	var totalUSD decimal.Decimal
	var entrySupplyMovementsResponses []entrySupplyMovementsResponse

	for i, supplyMovement := range entriesDomain {
		entrySupplyMovementsResponses = append(
			entrySupplyMovementsResponses,
			entrySupplyMovementsResponseFromDomain(supplyMovement),
	 	)
		totalUSD = totalUSD.Add(entrySupplyMovementsResponses[i].TotalUSD)
	}

	summary := summary{
		TotalKg: totalKg,
		TotalLt: totalLt,
		TotalUSD: totalUSD,
	}

	return GetEntrySupplyMovementsResponse{
		Summary: summary,
		EntrySupplyMovementsResponses: entrySupplyMovementsResponses,
	}
}


