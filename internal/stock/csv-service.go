package stock

import (
	"context"
	"strconv"

	"github.com/devpablocristo/ponti-backend/internal/shared/csvexport"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
)

// CSVExportFilename is the default download name.
const CSVExportFilename = "stock.csv"

const csvDateFormat = "02/01/2006"

var stockHeaders = []string{
	"INSUMO",
	"RUBRO",
	"INVERSOR",
	"INGRESADOS",
	"CONSUMIDOS",
	"STOCK",
	"STOCK REAL",
	"DIFERENCIA",
	"FECHA DE CIERRE",
	"PRECIO U.",
	"TOTAL U$",
}

// CSVExporter renders a list of Stock rows as a CSV byte slice.
type CSVExporter struct{}

func NewCSVExporter() *CSVExporter { return &CSVExporter{} }

func (e *CSVExporter) Export(_ context.Context, items []*domain.Stock) ([]byte, error) {
	rows := make([][]string, 0, len(items))
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

		closeDate := ""
		if it.CloseDate != nil {
			closeDate = it.CloseDate.Format(csvDateFormat)
		}

		rows = append(rows, []string{
			supplyName,
			classType,
			investorName,
			decToString(entry, 2),
			decToString(it.Consumed, 2),
			decToString(stockUnits, 2),
			decToString(it.RealStockUnits, 2),
			decToString(diff, 2),
			closeDate,
			decToString(unitPrice, 2),
			decToString(total, 2),
		})
	}
	return csvexport.Write(stockHeaders, rows)
}

func (e *CSVExporter) Close() error { return nil }

func decToString(d decimal.Decimal, scale int32) string {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return strconv.FormatFloat(f, 'f', int(scale), 64)
}
