package excel

import (
	"time"

	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
)

type StockExcelDTO struct {
	SupplyName      string     `excel:"INSUMO"`
	ClassType       string     `excel:"RUBRO"`
	InvestorName    string     `excel:"INVERSOR"`
	EntryStock      float64    `excel:"INGRESADOS"`
	Consumed        float64    `excel:"CONSUMIDOS"`
	StockUnits      float64    `excel:"STOCK"`
	RealStockUnits  float64    `excel:"STOCK REAL"`
	StockDifference float64    `excel:"DIFERENCIA"`
	CloseDate       *time.Time `excel:"FECHA DE CIERRE"`
	SupplyUnitPrice float64    `excel:"PRECIO U."`
	TotalUSD        float64    `excel:"TOTAL U$"`
}

func BuildExcelDTO(items []*domain.Stock) []StockExcelDTO {
	out := make([]StockExcelDTO, 0, len(items))

	for _, it := range items {
		entry := it.GetEntryStock()
		stockUnits := it.GetStockUnits()
		diff := decimal.Zero
if it.HasRealStockCount {
	diff = it.GetStockDifference()
}
		total := it.GetTotalUSD()

		investorName := ""
		if it.Investor != nil {
			investorName = it.Investor.Name
		}

		supplyName := ""
		classType := ""
		unitPrice := decimal.Zero
		if it.Supply != nil {
			supplyName = it.Supply.Name
			classType = it.Supply.CategoryName
			unitPrice = it.Supply.Price
		}

		out = append(out, StockExcelDTO{
			SupplyName:      supplyName,
			ClassType:       classType,
			InvestorName:    investorName,
			EntryStock:      decToFloat(entry, 2),
			Consumed:        decToFloat(it.Consumed, 2),
			StockUnits:      decToFloat(stockUnits, 2),
			RealStockUnits:  decToFloat(it.RealStockUnits, 2),
			StockDifference: decToFloat(diff, 2),
			CloseDate:       it.CloseDate,
			SupplyUnitPrice: decToFloat(unitPrice, 2),
			TotalUSD:        decToFloat(total, 2),
		})
	}
	return out
}

// helper para convertir decimal en float64 con redondeo opcional
func decToFloat(d decimal.Decimal, scale int32) float64 {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return f
}
