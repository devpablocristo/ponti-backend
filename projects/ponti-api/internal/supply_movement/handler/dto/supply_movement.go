package dto

// TODO: Adapt GetStocksResponse, GetStock for supply_movement context

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"time"
)

type SupplyMovementResponse struct {
	ID                   int64     `json:"id"`
	StockId              int64     `json:"stock_id"`
	Quantity             float64   `json:"quantity"`
	MovementType         string    `json:"movement_type"`
	MovementDate         time.Time `json:"movement_date"`
	Reference            string    `json:"reference"`
	ProjectId            int64     `json:"project_id"`
	ProjectDestinationId int64     `json:"project_destination_id"`
	SupplyID             int64     `json:"supply_id"`
	InvestorID           int64     `json:"investor_id"`
	ProviderID           int64     `json:"provider_id"`
}

type SupplyMovementsResponse struct {
	Items []SupplyMovementResponse `json:"items"`
}

func FromDomain(m *domain.SupplyMovement) SupplyMovementResponse {
	return SupplyMovementResponse{
		ID:                   m.ID,
		StockId:              m.StockId,
		Quantity:             m.Quantity,
		MovementType:         m.MovementType,
		MovementDate:         m.MovementDate,
		Reference:            m.Reference,
		ProjectId:            m.ProjectId,
		ProjectDestinationId: m.ProjectDestinationId,
		SupplyID:             m.Supply.ID,
		InvestorID:           m.Investor.ID,
		ProviderID:           m.Provider.ID,
	}
}

func FromDomainList(list []*domain.SupplyMovement) SupplyMovementsResponse {
	items := make([]SupplyMovementResponse, 0, len(list))
	for _, m := range list {
		items = append(items, FromDomain(m))
	}
	return SupplyMovementsResponse{Items: items}
}
