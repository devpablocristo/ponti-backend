package supply

import (
	"bytes"
	"context"
	"io"

	"github.com/devpablocristo/core/errors/go/domainerr"
	supplyExcel "github.com/devpablocristo/ponti-backend/internal/supply/excel"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type XLSXEnginePort interface {
	ExportToWriter(rows any, w io.Writer) error
	Close() error
}

type ExcelExporter struct {
	eng XLSXEnginePort
}

func NewExcelExporter(eng XLSXEnginePort) *ExcelExporter {
	return &ExcelExporter{eng: eng}
}

func (e *ExcelExporter) ExportSupplies(ctx context.Context, items []*domain.Supply) ([]byte, error) {
	_ = ctx

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	supplies := make([]domain.Supply, len(items))
	for i := range items {
		supplies[i] = *items[i]
	}

	rows := supplyExcel.BuildDTO(supplies)
	var buf bytes.Buffer
	if err := e.eng.ExportToWriter(rows, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ExcelExporter) ExportSupplyMovements(ctx context.Context, items []*domain.SupplyMovement) ([]byte, error) {
	_ = ctx

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	rows := supplyExcel.BuildSupplyMovementDTO(items)
	var buf bytes.Buffer
	if err := e.eng.ExportToWriter(rows, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ExcelExporter) Close() error { return e.eng.Close() }
