package workorder

import (
	"context"
	"strconv"

	"github.com/devpablocristo/ponti-backend/internal/shared/csvexport"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
	"github.com/shopspring/decimal"
)

// CSVExportFilename is the default download name.
const CSVExportFilename = "ordenes_de_trabajos.csv"

const csvDateFormat = "02/01/2006"

var workOrderHeaders = []string{
	"NUMERO DE ORDEN",
	"PROYECTO",
	"CAMPO",
	"LOTE",
	"FECHA",
	"CULTIVO",
	"LABOR",
	"TIPO/CLASE",
	"CONTRATISTA",
	"SUPERFICIE",
	"INSUMO",
	"CONSUMO",
	"RUBRO",
	"DOSIS",
	"COST U$/HA",
	"PRECIO UNIDAD",
	"TOTAL COSTO",
}

// CSVExporter renders a list of WorkOrderListElement rows as a CSV byte slice.
type CSVExporter struct{}

func NewCSVExporter() *CSVExporter { return &CSVExporter{} }

func (e *CSVExporter) Export(_ context.Context, items []domain.WorkOrderListElement) ([]byte, error) {
	rows := make([][]string, 0, len(items)+1)

	var totalSurfaceHa, totalConsumption, totalCost decimal.Decimal
	for _, it := range items {
		totalSurfaceHa = totalSurfaceHa.Add(it.SurfaceHa)
		totalConsumption = totalConsumption.Add(it.Consumption)
		totalCost = totalCost.Add(it.TotalCost)

		date := ""
		if !it.Date.IsZero() {
			date = it.Date.Format(csvDateFormat)
		}

		rows = append(rows, []string{
			it.Number,
			it.ProjectName,
			it.FieldName,
			it.LotName,
			date,
			it.CropName,
			it.LaborName,
			it.TypeName,
			it.Contractor,
			decToString(it.SurfaceHa, 2),
			it.SupplyName,
			decToString(it.Consumption, 2),
			it.CategoryName,
			decToString(it.Dose, 2),
			decToString(it.CostPerHa, 2),
			decToString(it.UnitPrice, 2),
			decToString(it.TotalCost, 2),
		})
	}

	if len(items) > 0 {
		rows = append(rows, []string{
			"TOTAL",
			"", "", "", "", "", "", "", "",
			decToString(totalSurfaceHa, 2),
			"",
			decToString(totalConsumption, 2),
			"",
			"0.00",
			"0.00",
			"0.00",
			decToString(totalCost, 2),
		})
	}

	return csvexport.Write(workOrderHeaders, rows)
}

func (e *CSVExporter) Close() error { return nil }

func decToString(d decimal.Decimal, scale int32) string {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return strconv.FormatFloat(f, 'f', int(scale), 64)
}
