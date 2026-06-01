# feature-013 · be-csv-export — implementation-status

## Estado global

- **En el SOURCE (`777e5f6a` / dp~1):** **completa y consistente.** Cero referencias residuales a `xuri/excelize`, `XLSXEnginePort`, `ExportToWriter`, `NewExcelExporter` en todo el árbol; wire regenerado (`wire_gen.go` usa `ProvideXExporterPort()` sin engine); tests de export migrados a CSV. La feature está terminada en su rama de origen.
- **En `develop` (destino):** **no portada (0%).** Hoy `develop` sigue con XLSX/excelize (esta extracción la introduce).
- **% completitud del refactor (en source):** ~100% del refactor XLSX->CSV.
- **% portabilidad limpia con el flist tal cual:** ~70%. Los 8 archivos del flist son extraíbles whole-file, pero la feature NO compila sin los archivos COMPARTIDOS (usecases/handler/wire/go.mod/excelize/mocks) que no están en el flist.

## Estado por componente

| componente | estado | nota |
|---|---|---|
| `csvexport.Write` | completa | autocontenido; falta test unitario propio (no hay `writer_test.go`) |
| `lot/csv-service.go` | completa | 18 columnas, fechas siembra/cosecha desde `it.Dates` |
| `labor/csv-service.go` | completa | `Export` (21 col) + `ExportTable` (5 col) |
| `stock/csv-service.go` | completa | usa getters `GetEntryStock`/`GetStockUnits`/`GetStockDifference`/`GetTotalUSD` |
| `supply/csv-service.go` | completa | `ExportSupplies` + `ExportSupplyMovements` |
| `work-order/csv-service.go` | completa | incluye fila TOTAL (dosis/cost/precio en TOTAL salen `"0.00"`, posible bug menor cosmético) |
| borrado Excel por dominio | completa | 17 archivos D |
| borrado engine `excelize` plataforma | completa (en source) | NO en flist — recordar borrarlo |
| wire providers | completa | hunks limpios |
| `wire_gen.go` | completa (en source) | regenerar al portar |
| `go.mod`/`go.sum` | completa (en source) | solo tomar la línea excelize |
| tests export | completa | `handler_export_test.go` valida `text/csv` |

## Tests

- **Existen y migrados a CSV:** `internal/lot/handler_export_test.go` (afirma `"csv"` y Content-Type `text/csv; charset=utf-8`), stubs en `labor`/`supply`/`work-order`.
- **Falta:** test unitario directo de `csvexport.Write` (BOM, `sep=;`, separador, NotFound con 0 filas). Recomendado agregar (mejora-futura, no bloqueante).

## Pendientes / bugs / dudas — clasificados

### BLOQUEANTE para mergear (en `develop`)
- Portar los archivos COMPARTIDOS no listados: 5 `usecases.go`, 5 `handler.go`, 5 `wire/*_providers.go`, `wire/wire.go`, `wire/wire_gen.go`, borrado engine `excelize`, edición `go.mod`/`go.sum`, mocks supply. Sin esto NO compila.
- Resolver import `domainerr` (core vs platform) en `writer.go` según el estado de `develop`.
- Regenerar `wire_gen.go` y `mock_repository.go` (no copiar literal).

### Mejora-futura
- `writer_test.go` unitario del paquete `csvexport`.
- Auditar paridad de columnas CSV vs el XLSX anterior (orden/labels) si QA lo requiere.

### Deuda-aceptable
- Fila TOTAL de work-order pone `"0.00"` en dosis/cost-u$/ha/precio-unidad: es intencional (no se suman), aceptable.
- Cada `csv-service.go` redefine su propio `decToString` (duplicación menor); aceptable.

### Duda-humana
- ¿Se actualiza el BFF FE en el mismo ciclo? (decisión de release; ver risks.md / cross-repo). Recomendado: BE primero, FE follow-up.
- ¿El repo tiene tooling de `wire`/`mockgen` disponible para regenerar? Revisar Makefile.
