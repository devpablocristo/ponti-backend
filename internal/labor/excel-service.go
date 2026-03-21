package labor

import (
	"bytes"
	"context"
	"io"

	"github.com/devpablocristo/core/backend/go/domainerr"
	labexcel "github.com/devpablocristo/ponti-backend/internal/labor/excel"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
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

func (e *ExcelExporter) Export(ctx context.Context, items []domain.LaborListItem) ([]byte, error) {
	_ = ctx
	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}
	rows := labexcel.BuildExcelDTO(items)
	var buf bytes.Buffer
	if err := e.eng.ExportToWriter(rows, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ExcelExporter) ExportTable(ctx context.Context, items []domain.ListedLabor) ([]byte, error) {
	_ = ctx
	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}
	rows := labexcel.BuildExcelTableDTO(items)
	var buf bytes.Buffer
	if err := e.eng.ExportToWriter(rows, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ExcelExporter) Close() error { return e.eng.Close() }
