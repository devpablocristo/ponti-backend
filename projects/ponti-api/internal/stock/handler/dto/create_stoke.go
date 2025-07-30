package dto

import (
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
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
	SupplyID     int64 `json:"supply_id"`
	InvestorID   int64 `json:"investor_id"`
	UnitsEntered int64 `json:"units_entered"`
	YearPeriod   int64 `json:"year_period"`
	MonthPeriod  int64 `json:"month_period"`
}

func (c CreateStock) ToDomain(projectID, fieldID int64, createdBy *int64) *domain.Stock {
	return &domain.Stock{
		Project:        &projdom.Project{ID: projectID},
		Field:          &fielddom.Field{ID: fieldID},
		Supply:         &supplydomain.Supply{ID: c.SupplyID},
		Investor:       &investordomain.Investor{ID: c.InvestorID},
		UnitsEntered:   c.UnitsEntered,
		YearPeriod:     c.YearPeriod,
		MonthPeriod:    c.MonthPeriod,
		RealStockUnits: c.UnitsEntered,
		Base: shareddomain.Base{
			CreatedBy: createdBy,
			UpdatedBy: createdBy,
		},
	}
}

func (cs *CreateStock) Validate() error {
	if cs.MonthPeriod == 0 {
		return types.NewMissingFieldError("month_period")
	}
	if cs.MonthPeriod < 0 {
		return types.NewValidationError("month_period", "must be greater than 0")
	}
	if cs.MonthPeriod > 12 {
		return types.NewValidationError("month_period", "must be less than or equal to 12")
	}
	if cs.YearPeriod == 0 {
		return types.NewMissingFieldError("year_period")
	}
	if cs.YearPeriod < 0 {
		return types.NewValidationError("year_period", "must be greater than or equal to 0")
	}
	if cs.SupplyID == 0 {
		return types.NewMissingFieldError("supply_id")
	}
	if cs.SupplyID < 0 {
		return types.NewValidationError("supply_id", "must be greater than or equal to 0")
	}
	if cs.InvestorID == 0 {
		return types.NewMissingFieldError("investor_id")
	}
	if cs.InvestorID < 0 {
		return types.NewValidationError("investor_id", "must be greater than or equal to 0")
	}
	return nil
}
