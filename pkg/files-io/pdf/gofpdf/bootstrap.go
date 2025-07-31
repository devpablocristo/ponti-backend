package pkgpdf

import (
	"os"
	"strconv"
)

// Bootstrap inicializa el repositorio PDF leyendo params o ENV (igual a pkggorm).
// ENV: PDF_FILE_PATH, PDF_ORIENTATION, PDF_PAGE_SIZE, PDF_MARGIN_LEFT, PDF_MARGIN_TOP,
// PDF_MARGIN_RIGHT, PDF_FONT_FAMILY, PDF_FONT_SIZE, PDF_TITLE
func Bootstrap(
	filePath, orientation, pageSize, fontFamily, title string,
	fontSize, ml, mt, mr float64,
	columnWidths []float64,
) (*Repository, error) {

	if filePath == "" {
		filePath = os.Getenv("PDF_FILE_PATH")
	}
	if orientation == "" {
		orientation = os.Getenv("PDF_ORIENTATION")
	}
	if pageSize == "" {
		pageSize = os.Getenv("PDF_PAGE_SIZE")
	}
	if fontFamily == "" {
		fontFamily = os.Getenv("PDF_FONT_FAMILY")
	}
	if title == "" {
		title = os.Getenv("PDF_TITLE")
	}
	if fontSize == 0 {
		if v := os.Getenv("PDF_FONT_SIZE"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				fontSize = f
			}
		}
	}
	if ml == 0 {
		if v := os.Getenv("PDF_MARGIN_LEFT"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				ml = f
			}
		}
	}
	if mt == 0 {
		if v := os.Getenv("PDF_MARGIN_TOP"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				mt = f
			}
		}
	}
	if mr == 0 {
		if v := os.Getenv("PDF_MARGIN_RIGHT"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				mr = f
			}
		}
	}

	cfg := newConfig(filePath, orientation, pageSize, fontFamily, title, fontSize, ml, mt, mr, columnWidths)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newRepository(cfg)
}
