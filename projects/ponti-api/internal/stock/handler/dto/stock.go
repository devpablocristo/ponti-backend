package dto

import (
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type GetStocksResponse struct {
	Stocks      []GetStock `json:"items"`
	NetTotalUSD float64    `json:"net_total_usd"`
}

type GetStock struct {
	ID              int64      `json:"id"`
	SupplyName      string     `json:"supply_name"`
	InvestorName    string     `json:"investor_name"`
	UnitsEntered    int64      `json:"units_entered"`
	UnitsConsumed   int64      `json:"units_consumed"`
	StockUnits      int64      `json:"stock_units"`
	RealStockUnits  int64      `json:"real_stock_units"`
	StockDifference int64      `json:"stock_difference"`
	TotalUSD        float64    `json:"total_usd"`
	CloseDate       *time.Time `json:"close_date"`
}

// FromDomain maps domain.Stock to GetStock DTO
func FromDomain(s *domain.Stock) *GetStock {
	return &GetStock{
		ID:              s.ID,
		InvestorName:    s.Investor.Name,
		SupplyName:      s.Supply.Name,
		UnitsEntered:    s.UnitsEntered,
		UnitsConsumed:   s.UnitsConsumed,
		StockUnits:      s.GetStockUnits(),
		RealStockUnits:  s.RealStockUnits,
		TotalUSD:        s.GetTotalUSD(),
		StockDifference: s.GetStockDifference(),
		CloseDate:       s.CloseDate,
	}
}

func NewGetStocksListed(stocks []*domain.Stock) GetStocksResponse {
	var netTotalUSD float64
	listedStocks := make([]GetStock, len(stocks))

	for i, stock := range stocks {
		listedStocks[i] = *FromDomain(stock)
		netTotalUSD += stock.GetTotalUSD()
	}

	return GetStocksResponse{
		Stocks:      listedStocks,
		NetTotalUSD: netTotalUSD,
	}

}
