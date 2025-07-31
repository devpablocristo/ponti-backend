package pkgexcel

import (
	"os"
	"strconv"
)

// Bootstrap inicializa el repositorio Excel leyendo params directos o ENV (igual a pkggorm)
// ENV soportadas: EXCEL_FILE_PATH, EXCEL_SHEET, EXCEL_DATE_FORMAT, EXCEL_WRITE_HEADER, EXCEL_COLUMN_WIDTHS (no implementado parse widths aquí)
func Bootstrap(filePath, sheet, dateFormat string, writeHeader bool, columnWidths []float64) (*Service, error) {
	if filePath == "" {
		filePath = os.Getenv("EXCEL_FILE_PATH")
	}
	if sheet == "" {
		sheet = os.Getenv("EXCEL_SHEET")
	}
	if dateFormat == "" {
		if v := os.Getenv("EXCEL_DATE_FORMAT"); v != "" { dateFormat = v }
	}
	// writeHeader por defecto true si no viene explícito
	if !writeHeader {
		if v := os.Getenv("EXCEL_WRITE_HEADER"); v != "" {
			if b, err := strconv.ParseBool(v); err == nil { writeHeader = b }
		} else {
			writeHeader = true
		}
	}

	cfg := newConfig(filePath, sheet, dateFormat, writeHeader, columnWidths)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newService(cfg)
}



