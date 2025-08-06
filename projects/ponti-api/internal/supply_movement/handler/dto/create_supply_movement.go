package dto

// TODO: Adapt CreateStocksRequest, CreateStocksResponse, CreateStock for supply_movement context

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"time"
)

type CreateSupplyMovementRequest struct {
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

func (r *CreateSupplyMovementRequest) ToDomain() *domain.SupplyMovement {
	return &domain.SupplyMovement{
		StockId:              r.StockId,
		Quantity:             r.Quantity,
		MovementType:         r.MovementType,
		MovementDate:         r.MovementDate,
		Reference:            r.Reference,
		ProjectId:            r.ProjectId,
		ProjectDestinationId: r.ProjectDestinationId,
		Supply:               &domain.Supply{ID: r.SupplyID},
		Investor:             &domain.Investor{ID: r.InvestorID},
		Provider:             &domain.Provider{ID: r.ProviderID},
	}
}
