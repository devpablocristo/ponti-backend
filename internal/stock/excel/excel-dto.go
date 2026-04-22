package excel

import (
	"time"

	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
)

type StockExcelDTO struct {
	SupplyName      string     `excel:"INSUMO"`
	ClassType       string     `excel:"RUBRO"`
	EntryStock      float64    `excel:"INGRESADOS"`
	OutStock        float64    `excel:"SALIDOS"`
	Consumed        float64    `excel:"CONSUMIDOS"`
	StockUnits      float64    `excel:"STOCK SISTEMA"`
	RealStockUnits  float64    `excel:"STOCK FISICO"`
	StockDifference float64    `excel:"DIFERENCIA"`
	LastCountAt     *time.Time `excel:"ULTIMO CONTEO"`
	SupplyUnitPrice float64    `excel:"PRECIO U."`
	TotalUSD        float64    `excel:"TOTAL U$"`
}

func BuildExcelDTO(items []*domain.Stock) []StockExcelDTO {
	out := make([]StockExcelDTO, 0, len(items))

	for _, it := range items {
		diff := decimal.Zero
		if it.HasRealStockCount {
			diff = it.GetStockDifference()
		}
		total := it.GetTotalUSD()

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
			EntryStock:      decToFloat(it.EntryStock, 2),
			OutStock:        decToFloat(it.OutStock, 2),
			Consumed:        decToFloat(it.Consumed, 2),
			StockUnits:      decToFloat(it.GetStockUnits(), 2),
			RealStockUnits:  decToFloat(it.RealStockUnits, 2),
			StockDifference: decToFloat(diff, 2),
			LastCountAt:     it.LastCountAt,
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
