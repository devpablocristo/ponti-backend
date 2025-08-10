package dto

import (
	"time"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type GetStocksResponse struct {
	Stocks      []GetStockSummary `json:"items"`
	NetTotalUSD float64    `json:"net_total_usd"`
	TotalLiters float32 `json:"total_liters"`
	TotalKilograms float32 `json:"total_kilograms"`
}

type GetStockSummary struct {
	ID              int64      `json:"id"`
	SupplyName      string     `json:"supply_name"`
	InvestorName    string     `json:"investor_name"`
	StockUnits      float64      `json:"stock_units"`
	RealStockUnits  float64      `json:"real_stock_units"`
	StockDifference float64      `json:"stock_difference"`
	TotalUSD        float64    `json:"total_usd"`
	ClassType       string     `json:"class_type"`
	CloseDate       *time.Time `json:"close_date"`

	supplyUnitId int64

}

// FromDomain maps domain.Stock to GetStock DTO
func FromDomain(s *domain.Stock) *GetStockSummary {
	return &GetStockSummary{
		ID:              s.ID,
		InvestorName:    s.Investor.Name,
		SupplyName:      s.Supply.Name,
		StockUnits:      s.GetStockUnits(),
		RealStockUnits:  s.RealStockUnits,
		TotalUSD:        s.GetTotalUSD(),
		StockDifference: s.GetStockDifference(),
		CloseDate:       s.CloseDate,
		ClassType:       s.Supply.Type.Name,
		supplyUnitId: s.Supply.UnitID,

	}
}

func NewGetStocksListed(stocks []*domain.Stock) GetStocksResponse {
	var netTotalUSD float64
	var totalKilograms float32
	var totalLiters float32
	listedStocks := make([]GetStockSummary, len(stocks))

	for i, stock := range stocks {
		listedStocks[i] = *FromDomain(stock)
		netTotalUSD += stock.GetTotalUSD()
		if listedStocks[i].supplyUnitId == 1 {
			totalKilograms =+ float32(stock.GetStockUnits())
		}
		if listedStocks[i].supplyUnitId == 2 {
			totalLiters =+ float32(stock.GetStockUnits())
		}
	}



	return GetStocksResponse{
		Stocks:      listedStocks,
		NetTotalUSD: netTotalUSD,
		TotalLiters: totalLiters,
		TotalKilograms: totalKilograms,
	}

}
