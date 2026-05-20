package labor

import (
	"context"
	"strconv"

	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/csvexport"
	"github.com/shopspring/decimal"
)

// CSVExportFilename / CSVTableExportFilename are the default download names.
const (
	CSVExportFilename      = "labores.csv"
	CSVTableExportFilename = "labores_tabla.csv"
)

const csvDateFormat = "02/01/2006"

var laborProjectHeaders = []string{
	"OT N°",
	"FECHA",
	"PROYECTO",
	"CAMPO",
	"CULTIVO",
	"LABOR",
	"CONTRATISTA",
	"SUPÉRFICIE",
	"COSTO $/HECTÁREA",
	"INVERSOR",
	"U$ PROM",
	"TOTAL $ / NETO",
	"TOTAL $ / IVA",
	"COSTO U$ /HA",
	"TOTAL U$ NETO",
	"N° FACTURA",
	"EMPRESA",
	"FECHA DE FACTURACIÓN",
	"ESTADO DE FACTURA",
}

var laborTableHeaders = []string{
	"NOMBRE",
	"CATEGORÍA",
	"PRECIO",
	"ESTADO_PRECIO",
	"CONTRATISTA",
}

// CSVExporter exposes both Export (project view) and ExportTable (database view).
type CSVExporter struct{}

func NewCSVExporter() *CSVExporter { return &CSVExporter{} }

func (e *CSVExporter) Export(_ context.Context, items []domain.LaborListItem) ([]byte, error) {
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		date := ""
		if !it.Date.IsZero() {
			date = it.Date.Format(csvDateFormat)
		}
		invoiceDate := ""
		if it.InvoiceDate != nil {
			invoiceDate = it.InvoiceDate.Format(csvDateFormat)
		}
		rows = append(rows, []string{
			it.WorkOrderNumber,
			date,
			it.ProjectName,
			it.FieldName,
			it.CropName,
			it.LaborName,
			it.Contractor,
			decToString(it.SurfaceHa, 2),
			decToString(it.CostHa, 2),
			it.InvestorName,
			decToString(it.USDAvgValue, 2),
			decToString(it.NetTotal, 2),
			decToString(it.TotalIVA, 2),
			decToString(it.USDCostHa, 2),
			decToString(it.USDNetTotal, 2),
			it.InvoiceNumber,
			it.InvoiceCompany,
			invoiceDate,
			it.InvoiceStatus,
		})
	}
	return csvexport.Write(laborProjectHeaders, rows)
}

func (e *CSVExporter) ExportTable(_ context.Context, items []domain.ListedLabor) ([]byte, error) {
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		priceStatus := "Final"
		if it.IsPartialPrice {
			priceStatus = "Parcial"
		}
		rows = append(rows, []string{
			it.Name,
			it.CategoryName,
			decToString(it.Price, 2),
			priceStatus,
			it.ContractorName,
		})
	}
	return csvexport.Write(laborTableHeaders, rows)
}

func (e *CSVExporter) Close() error { return nil }

func decToString(d decimal.Decimal, scale int32) string {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return strconv.FormatFloat(f, 'f', int(scale), 64)
}
