package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type GetStockByIdResponse struct {
	ID             int64      `json:"id"`
	ProjectID      int64      `json:"project_id"`
	FieldID        int64      `json:"field_id"`
	SupplyID       int64      `json:"supply_id"`
	InvestorID     int64      `json:"investor_id"`
	ClassType      string     `json:"class_type"`
	CloseDate      *time.Time `json:"close_date"`
	UnitsEntered   int64      `json:"units_entered"`
	UnitsConsumed  int64      `json:"units_consumed"`
	RealStockUnits int64      `json:"real_stock_units"`
}

func StockByIdResponseFromDomain(stock *domain.Stock) *GetStockByIdResponse {
	return &GetStockByIdResponse{
		ID:             stock.ID,
		ProjectID:      stock.Project.ID,
		FieldID:        stock.Field.ID,
		SupplyID:       stock.Supply.ID,
		InvestorID:     stock.Investor.ID,
		CloseDate:      stock.CloseDate,
		UnitsEntered:   stock.UnitsEntered,
		UnitsConsumed:  stock.UnitsConsumed,
		RealStockUnits: stock.RealStockUnits,
		ClassType:      stock.Supply.Type.Name,
	}
}
