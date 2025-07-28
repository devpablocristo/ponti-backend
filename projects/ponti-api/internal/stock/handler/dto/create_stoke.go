package dto

import (
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	projdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	supplydomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/usecases/domain"
)

type CreateStocksRequest struct {
	Stocks []CreateStock `json:"stocks"`
}

type CreateStocksResponse struct {
	StockID     int64  `json:"stock_id"`
	SupplyID    int64  `json:"supply_id"`
	IsSaved     bool   `json:"is_saved"`
	ErrorDetail string `json:"error_detail"`
}
type CreateStock struct {
	ProjectID    int64  `json:"project_id"`
	FieldID      int64  `json:"field_id"`
	SupplyID     int64  `json:"supply_id"`
	InvestorID   int64  `json:"investor_id"`
	UnitsEntered int64  `json:"units_entered"`
	CreatedBy    *int64 `json:"created_by"`
	UpdatedBy    *int64 `json:"updated_by"`
}

func (c CreateStock) ToDomain() *domain.Stock {
	return &domain.Stock{
		Project:      &projdom.Project{ID: c.ProjectID},
		Field:        &fielddom.Field{ID: c.FieldID},
		Supply:       &supplydomain.Supply{ID: c.SupplyID},
		Investor:     &investordomain.Investor{ID: c.InvestorID},
		UnitsEntered: c.UnitsEntered,
		Base: shareddomain.Base{
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
		},
	}
}
