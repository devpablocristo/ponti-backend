package dto

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type GetStocksResponse struct {
	Stocks         []GetStockSummary `json:"items"`
	NetTotalUSD    decimal.Decimal   `json:"net_total_usd"`
	TotalLiters    decimal.Decimal   `json:"total_liters"`
	TotalKilograms decimal.Decimal   `json:"total_kilograms"`
}

type GetStockSummary struct {
	ID              int64           `json:"id"`
	SupplyName      string          `json:"supply_name"`
	InvestorName    string          `json:"investor_name"`
	StockUnits      decimal.Decimal `json:"stock_units"`
	RealStockUnits  decimal.Decimal `json:"real_stock_units"`
	StockDifference decimal.Decimal `json:"stock_difference"`
	TotalUSD        decimal.Decimal `json:"total_usd"`
	ClassType       string          `json:"class_type"`
	CloseDate       *time.Time      `json:"close_date"`
	SupplyUnitId    int64           `json:"supply_unit_id"`
	SupplyUnitPrice decimal.Decimal `json:"supply_unit_price"`
	EntryStock      decimal.Decimal `json:"entry_stock"`
	OutStock        decimal.Decimal `json:"out_stock"`
	Consumed        decimal.Decimal `json:"consumed"`
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
		SupplyUnitId:    s.Supply.UnitID,
		SupplyUnitPrice: s.Supply.Price,
		EntryStock:      s.GetEntryStock(),
		OutStock:        s.GetOutStock(),
		Consumed:        s.Consumed,
	}
}

func NewGetStocksListed(stocks []*domain.Stock) GetStocksResponse {
	var netTotalUSD decimal.Decimal
	var totalKilograms decimal.Decimal
	var totalLiters decimal.Decimal
	listedStocks := make([]GetStockSummary, len(stocks))

	for i, stock := range stocks {
		listedStocks[i] = *FromDomain(stock)
		netTotalUSD = netTotalUSD.Add(stock.GetTotalUSD())
		if isKG(stock.GetSupplyUnitName()) {
			totalKilograms = totalKilograms.Add(stock.GetStockUnits())
		} else if isLt(stock.GetSupplyUnitName()) {
			totalLiters = totalLiters.Add(stock.GetStockUnits())
		}
	}

	return GetStocksResponse{
		Stocks:         listedStocks,
		NetTotalUSD:    netTotalUSD,
		TotalLiters:    totalLiters,
		TotalKilograms: totalKilograms,
	}

}

func isKG(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "kg")
}

func isLt(unitName string) bool {
	return strings.Contains(strings.ToLower(unitName), "lt")
}
