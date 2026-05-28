package supply

import (
	"context"
	"strconv"

	"github.com/devpablocristo/ponti-backend/internal/shared/csvexport"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

// Default download filenames.
const (
	CSVSupplyTableFilename     = "insumos.csv"
	CSVSupplyMovementsFilename = "movimientos_insumos.csv"
)

const csvDateFormat = "02/01/2006"

var supplyTableHeaders = []string{
	"NOMBRE",
	"UNIDAD",
	"PRECIO",
	"ESTADO PRECIO",
	"RUBRO",
	"TIPO/CLASE",
}

var supplyMovementHeaders = []string{
	"INGRESO",
	"N° REMITO",
	"FECHA",
	"INVERSOR",
	"INSUMO",
	"CANTIDAD",
	"UNIDAD",
	"RUBRO",
	"TIPO/CLASE",
	"PROVEEDOR",
	"PRECIO U$",
	"TOTAL U$",
}

// CSVExporter renders supplies and supply movements as CSV byte slices.
type CSVExporter struct{}

func NewCSVExporter() *CSVExporter { return &CSVExporter{} }

func (e *CSVExporter) ExportSupplies(_ context.Context, items []*domain.Supply) ([]byte, error) {
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		priceStatus := "Final"
		if it.IsPartialPrice {
			priceStatus = "Parcial"
		}
		rows = append(rows, []string{
			it.Name,
			it.UnitName,
			decToString(it.Price, 2),
			priceStatus,
			it.CategoryName,
			it.Type.Name,
		})
	}
	return csvexport.Write(supplyTableHeaders, rows)
}

func (e *CSVExporter) ExportSupplyMovements(_ context.Context, items []*domain.SupplyMovement) ([]byte, error) {
	rows := make([][]string, 0, len(items))
	for _, it := range items {
		price, _ := it.Supply.Price.Float64()
		total, _ := it.Supply.Price.Mul(it.Quantity).Float64()
		movementDate := ""
		if it.MovementDate != nil {
			movementDate = it.MovementDate.Format(csvDateFormat)
		}
		rows = append(rows, []string{
			it.MovementType,
			it.ReferenceNumber,
			movementDate,
			it.Investor.Name,
			it.Supply.Name,
			decToString(it.Quantity, 2),
			it.Supply.UnitName,
			it.Supply.CategoryName,
			it.Supply.Type.Name,
			it.Provider.Name,
			strconv.FormatFloat(price, 'f', 2, 64),
			strconv.FormatFloat(total, 'f', 2, 64),
		})
	}
	return csvexport.Write(supplyMovementHeaders, rows)
}

func (e *CSVExporter) Close() error { return nil }

func decToString(d decimal.Decimal, scale int32) string {
	if scale >= 0 {
		d = d.Round(scale)
	}
	f, _ := d.Float64()
	return strconv.FormatFloat(f, 'f', int(scale), 64)
}
