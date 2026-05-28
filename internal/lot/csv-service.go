package lot

import (
	"context"
	"strconv"

	"github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/csvexport"
	"github.com/shopspring/decimal"
)

// CSVExportFilename is the default download name.
const CSVExportFilename = "lotes.csv"

var lotHeaders = []string{
	"PROYECTO",
	"CAMPO",
	"LOTES",
	"CULTIVO ANTERIOR",
	"CULTIVO ACTUAL",
	"VARIEDAD",
	"SUP. SIEMBRA",
	"FECHA SIEMBRA",
	"COSTO U$/HA",
	"SUP. COSECHA",
	"FECHA COSECHA",
	"TONELADAS",
	"RENDIMIENTO",
	"INGRESO NETO/HA",
	"ARRIENDO/HA",
	"ADMIN. PROYECTO/HA",
	"ACTIVO TOTAL/HA",
	"RESULTADO OPERATIVO",
}

const csvDateFormat = "02/01/2006"

// CSVExporter renders a list of LotTable rows as a CSV byte slice.
type CSVExporter struct{}

func NewCSVExporter() *CSVExporter { return &CSVExporter{} }

func (e *CSVExporter) Export(_ context.Context, items []domain.LotTable) ([]byte, error) {
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		sowing := ""
		harvest := ""
		for _, dt := range it.Dates {
			if sowing == "" && dt.SowingDate != nil {
				sowing = dt.SowingDate.Format(csvDateFormat)
			}
			if harvest == "" && dt.HarvestDate != nil {
				harvest = dt.HarvestDate.Format(csvDateFormat)
			}
			if sowing != "" && harvest != "" {
				break
			}
		}
		rows = append(rows, []string{
			it.ProjectName,
			it.FieldName,
			it.LotName,
			it.PreviousCrop,
			it.CurrentCrop,
			it.Variety,
			decToString(it.SowedArea, 2),
			sowing,
			decToString(it.CostUsdPerHa, 2),
			decToString(it.HarvestedArea, 2),
			harvest,
			decToString(it.Tons, 2),
			decToString(it.YieldTnPerHa, 2),
			decToString(it.IncomeNetPerHa, 2),
			decToString(it.RentPerHa, 2),
			decToString(it.AdminCost, 2),
			decToString(it.ActiveTotalPerHa, 2),
			decToString(it.OperatingResultPerHa, 2),
		})
	}
	return csvexport.Write(lotHeaders, rows)
}

func (e *CSVExporter) Close() error { return nil }

func decToString(d decimal.Decimal, scale int32) string {
	if scale >= 0 {
		d = d.Round(scale)
	}
	return strconv.FormatFloat(mustFloat(d), 'f', int(scale), 64)
}

func mustFloat(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}
