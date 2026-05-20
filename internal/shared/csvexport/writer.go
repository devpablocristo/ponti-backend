// Package csvexport writes data tables as CSV bytes that open cleanly in
// Excel / Google Sheets — UTF-8 with BOM, `;` separator, RFC 4180 quoting.
//
// CSV is the only export/import format used in the project. Every entity
// (lots, work orders, supply movements, etc.) calls Write to emit its rows.
package csvexport

import (
	"bytes"
	"encoding/csv"

	"github.com/devpablocristo/core/errors/go/domainerr"
)

const (
	// ContentType is the canonical MIME type for the exported file.
	ContentType = "text/csv; charset=utf-8"
	// Separator is used by Excel in Spanish locale. Recipients in other
	// locales can rely on the `sep=;` hint emitted before the header row.
	Separator = ';'
)

// utf8BOM tells Excel that the file is UTF-8 so it stops mangling accents.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Write returns a UTF-8 encoded CSV with BOM. Headers go first, then rows in
// the order provided. Returns NotFound if there are zero rows so callers can
// surface "nothing to export".
func Write(headers []string, rows [][]string) ([]byte, error) {
	if len(rows) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	var buf bytes.Buffer
	buf.Write(utf8BOM)
	// Tell Excel which separator we used; harmless in other tools.
	buf.WriteString("sep=;\n")

	w := csv.NewWriter(&buf)
	w.Comma = Separator
	if err := w.Write(headers); err != nil {
		return nil, domainerr.Internal("write csv header")
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return nil, domainerr.Internal("write csv row")
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, domainerr.Internal("flush csv")
	}
	return buf.Bytes(), nil
}
