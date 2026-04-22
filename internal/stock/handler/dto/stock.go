// Package dto define los DTOs HTTP para stock.
package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	domain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

type GetStocksResponse struct {
	Stocks         []GetStockSummary `json:"items"`
	NetTotalUSD    decimal.Decimal   `json:"net_total_usd"`
	TotalLiters    decimal.Decimal   `json:"total_liters"`
	TotalKilograms decimal.Decimal   `json:"total_kilograms"`
}

// MarshalJSON aplica redondeo: Total U$/Neto, Total insumos (kg), Total insumos (lts) al entero más próximo.
func (r GetStocksResponse) MarshalJSON() ([]byte, error) {
	aux := struct {
		Stocks         []GetStockSummary `json:"items"`
		NetTotalUSD    string            `json:"net_total_usd"`
		TotalLiters    string            `json:"total_liters"`
		TotalKilograms string            `json:"total_kilograms"`
	}{
		Stocks:         r.Stocks,
		NetTotalUSD:    r.NetTotalUSD.StringFixed(0),
		TotalLiters:    r.TotalLiters.StringFixed(0),
		TotalKilograms: r.TotalKilograms.StringFixed(0),
	}
	return json.Marshal(aux)
}

type GetStockSummary struct {
	SupplyID          int64            `json:"supply_id"`
	SupplyName        string           `json:"supply_name"`
	StockUnits        decimal.Decimal  `json:"stock_units"`
	RealStockUnits    decimal.Decimal  `json:"real_stock_units"`
	StockDifference   *decimal.Decimal `json:"stock_difference"`
	TotalUSD          decimal.Decimal  `json:"total_usd"`
	ClassType         string           `json:"class_type"`
	SupplyUnitID      int64            `json:"supply_unit_id"`
	SupplyUnitPrice   decimal.Decimal  `json:"supply_unit_price"`
	EntryStock        decimal.Decimal  `json:"entry_stock"`
	OutStock          decimal.Decimal  `json:"out_stock"`
	Consumed          decimal.Decimal  `json:"consumed"`
	LastCountAt       *time.Time       `json:"last_count_at"`
	HasRealStockCount bool             `json:"has_real_stock_count"`
}

// MarshalJSON aplica redondeo: Precio u: 2 dec, Total u$s: 2 dec.
func (s GetStockSummary) MarshalJSON() ([]byte, error) {
	var stockDifference any
	if s.StockDifference != nil {
		stockDifference = s.StockDifference.StringFixed(2)
	}

	aux := struct {
		SupplyID          int64      `json:"supply_id"`
		SupplyName        string     `json:"supply_name"`
		StockUnits        string     `json:"stock_units"`
		RealStockUnits    string     `json:"real_stock_units"`
		StockDifference   any        `json:"stock_difference"`
		TotalUSD          string     `json:"total_usd"`
		ClassType         string     `json:"class_type"`
		SupplyUnitID      int64      `json:"supply_unit_id"`
		SupplyUnitPrice   string     `json:"supply_unit_price"`
		EntryStock        string     `json:"entry_stock"`
		OutStock          string     `json:"out_stock"`
		Consumed          string     `json:"consumed"`
		LastCountAt       *time.Time `json:"last_count_at"`
		HasRealStockCount bool       `json:"has_real_stock_count"`
	}{
		SupplyID:          s.SupplyID,
		SupplyName:        s.SupplyName,
		StockUnits:        s.StockUnits.StringFixed(2),
		RealStockUnits:    s.RealStockUnits.StringFixed(2),
		StockDifference:   stockDifference,
		TotalUSD:          s.TotalUSD.StringFixed(2),
		ClassType:         s.ClassType,
		SupplyUnitID:      s.SupplyUnitID,
		SupplyUnitPrice:   s.SupplyUnitPrice.StringFixed(2),
		EntryStock:        s.EntryStock.StringFixed(2),
		OutStock:          s.OutStock.StringFixed(2),
		Consumed:          s.Consumed.StringFixed(2),
		LastCountAt:       s.LastCountAt,
		HasRealStockCount: s.HasRealStockCount,
	}
	return json.Marshal(aux)
}

// FromDomain maps domain.Stock to GetStock DTO.
func FromDomain(s *domain.Stock) *GetStockSummary {
	classType := ""
	supplyName := ""
	supplyUnitID := int64(0)
	supplyUnitPrice := decimal.Zero

	if s != nil && s.Supply != nil {
		classType = s.Supply.CategoryName
		supplyName = s.Supply.Name
		supplyUnitID = s.Supply.UnitID
		supplyUnitPrice = s.Supply.Price
	}

	return &GetStockSummary{
		SupplyID:          s.ID,
		SupplyName:        supplyName,
		StockUnits:        s.GetStockUnits(),
		RealStockUnits:    s.RealStockUnits,
		TotalUSD:          s.GetTotalUSD(),
		StockDifference:   s.GetStockDifferencePtr(),
		ClassType:         classType,
		SupplyUnitID:      supplyUnitID,
		SupplyUnitPrice:   supplyUnitPrice,
		EntryStock:        s.EntryStock,
		OutStock:          s.OutStock,
		Consumed:          s.Consumed,
		LastCountAt:       s.LastCountAt,
		HasRealStockCount: s.HasRealStockCount,
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
