// Package workorder contiene servicios de exportación para work orders.
package workorder

import (
	"bytes"
	"context"
	"io"

	workOrderExcel "github.com/devpablocristo/ponti-backend/internal/work-order/excel"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
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

func (e *ExcelExporter) Export(ctx context.Context, items []domain.WorkOrderListElement) ([]byte, error) {
	_ = ctx
	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	rows := workOrderExcel.BuildWorkOrderExcelDTO(items)
	var buf bytes.Buffer
	if err := e.eng.ExportToWriter(rows, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ExcelExporter) Close() error { return e.eng.Close() }
