# feature-013 · be-csv-export — spec

- **id:** feature-013
- **slug:** be-csv-export
- **nombre:** CSV export (replace XLSX)
- **tipo:** refactor
- **repo:** Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **merge:** BE-first
- **existe-en-FE:** No hay carpeta de feature en FE. La feature es solo-BE. Sin embargo SÍ hay impacto de contrato sobre el BFF de FE (`web/api/src/routes/*.ts`) — ver "alcance en el otro repo".
- **existe-en-BE:** Sí, es el repo principal.
- **SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`, merge de PR #120 quick-fix). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama destino:** `develop` (tip `003a9b8f`).
- **rango fuente-de-verdad (diff):** `0972e565..777e5f6a`.

## Resumen

Reemplaza por completo el motor de exportación XLSX (basado en `github.com/xuri/excelize/v2`) por un exportador CSV minimalista y compartido. Se introduce el paquete `internal/shared/csvexport` (un único `Write(headers, rows)`), y cada dominio que exportaba a Excel (`labor`, `lot`, `stock`, `supply`, `work-order`) reemplaza su `excel-service.go` + subpaquete `excel/` por un `csv-service.go` autocontenido. Se elimina el engine Excel de plataforma (`internal/platform/files/excel/excelize/*`) y la dependencia `xuri/excelize/v2` de `go.mod`/`go.sum`.

## Objetivo

- CSV como único formato de export/import del proyecto (lo dice el doc del propio paquete `csvexport`).
- Quitar la dependencia pesada `excelize` y todo el andamiaje de DTOs/config Excel por dominio.
- Simplificar la cadena de wiring: de `ProvideXPkgExcelService -> ProvideXXLSXEnginePort -> ProvideXExporterPort(eng)` a un único `ProvideXExporterPort()` sin engine.

## Problema (qué resolvía)

El export XLSX requería: un engine de plataforma (`pkgexcel.Bootstrap` con archivo temporal en `os.TempDir()`), una interfaz `XLSXEnginePort{ExportToWriter, Close}`, un DTO Excel por dominio (`Build...ExcelDTO`), config de hojas/anchos de columna, y 3 providers de wire por dominio. CSV elimina todo eso: el exportador construye `[][]string` y delega en `csvexport.Write`.

## Alcance en este repo (BE)

Archivos del flist (creados/borrados) — núcleo de la feature:

- **NUEVO** `internal/shared/csvexport/writer.go` — paquete compartido. `Write([]string headers, [][]string rows) ([]byte, error)`. UTF-8 con BOM, separador `;`, prefijo `sep=;\n` para Excel, quoting RFC 4180. Constantes `ContentType = "text/csv; charset=utf-8"` y `Separator = ';'`. Devuelve `domainerr.NotFound("there is no data to export")` si `len(rows)==0`.
- **NUEVO** `internal/labor/csv-service.go` — `CSVExporter` con `Export` (vista proyecto, 21 columnas) y `ExportTable` (vista base de datos, 5 columnas). Filenames `labores.csv` / `labores_tabla.csv`.
- **NUEVO** `internal/lot/csv-service.go` — `CSVExporter.Export([]domain.LotTable)`. 18 columnas. Filename `lotes.csv`.
- **NUEVO** `internal/stock/csv-service.go` — `CSVExporter.Export([]*domain.Stock)`. 11 columnas. Filename `stock.csv`.
- **NUEVO** `internal/supply/csv-service.go` — `CSVExporter` con `ExportSupplies` (6 col) y `ExportSupplyMovements` (12 col). Filenames `insumos.csv` / `movimientos_insumos.csv`.
- **NUEVO** `internal/work-order/csv-service.go` — `CSVExporter.Export([]domain.WorkOrderListElement)`. 18 columnas + fila TOTAL al final. Filename `ordenes_de_trabajos.csv`.
- **BORRADOS** por dominio: `excel-service.go`, y subpaquete `excel/` (`config.go`, `excel-dto.go`, `excel-table-dto.go`, `excel_dto.go`, `excel_dto_table.go`, `helpers.go` según dominio).

Archivos NO en el flist pero IMPRESCINDIBLES para que compile (COMPARTIDOS / partial-hunks — ver file-list.md):

- `internal/{labor,lot,stock,supply,work-order}/usecases.go` — rename del campo/param `excel ExporterAdapterPort` -> `exporter ExporterAdapterPort`; `u.excel.Export(...)` -> `u.exporter.Export(...)`. **Hunks mezclados** con otras features (lot-metrics `UpdateLotTons`, etc.).
- `internal/{labor,lot,stock,supply,work-order}/handler.go` — import `lotExcel`/etc -> `csvexport`; `Content-Type` de `application/vnd.openxmlformats-...sheet` -> `csvexport.ContentType`; filename `DefaultFilename` (.xlsx) -> `CSVExportFilename` (.csv); se agrega `c.Data(http.StatusOK, csvexport.ContentType, data)`. **Hunks mezclados** con otras features.
- `wire/{labor,lot,stock,supply,work_order}_providers.go` — eliminan `ProvideXPkgExcelService`, `ProvideXXLSXEnginePort`, `LotExcelService` struct, y simplifican `ProvideXExporterPort()`; quitan providers del `wire.NewSet`. **Hunks dedicados y limpios** a esta feature.
- `wire/wire.go` y `wire/wire_gen.go` — código generado/registro; las líneas de Excel se eliminan. **Compartido y de alto riesgo** (mezcla todas las features).
- **BORRADOS** `internal/platform/files/excel/excelize/{bootstrap.go,config.go,service.go}` — engine Excel de plataforma (NO está en el flist, pero es parte de la feature).
- `go.mod` / `go.sum` — quitan `github.com/xuri/excelize/v2 v2.9.1`. **Muy mezclados** con la migración core->platform (new-cns3): NO copiar enteros.
- `internal/supply/mocks/mock_repository.go` — el mock de `MockExporterAdapterPort` debe exponer `ExportSupplies`/`ExportSupplyMovements` (sin métodos XLSX). **Compartido / regenerado**.
- Tests de export: `internal/lot/handler_export_test.go` (`"xlsx"`->`"csv"`, Content-Type esperado), stubs en `internal/{labor,supply}/handler_*_test.go`.

## Alcance en el otro repo (FE / web)

La consigna dice "Solo-BE: sin cambios FE". Es correcto en cuanto a que no hay carpeta de feature FE. PERO el BFF de FE (`/home/pablocristo/Proyectos/pablo/ponti/web/api/src/routes/*.ts`) hoy **hardcodea** la respuesta del proxy de export como XLSX. Ejemplo en `lots.ts`:

```
res.setHeader("Content-Disposition", 'attachment; filename="lots.xlsx"');
res.setHeader("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet");
```

Tras este refactor el BE devuelve bytes CSV (`text/csv`) pero el BFF los re-envuelve como `.xlsx` -> el usuario descargaría un archivo `.xlsx` que en realidad es CSV (Excel suele abrirlo igual gracias al BOM y `sep=;`, pero el nombre/MIME quedan inconsistentes). Archivos FE afectados: `web/api/src/routes/{lots,labors,workorders,stock,stock_movements,movements}.ts`. **Esto NO se arregla en feature-013** (es solo-BE); se documenta como follow-up FE (encaja en el área de 014/017 o un fix de BFF aparte).

## Fuera de alcance

- Ajustar el BFF FE (otra feature/PR FE).
- La migración core->platform de `go.mod`/`go.sum` (es de new-cns3, NO de 013) — solo se toma la línea de `excelize`.
- `UpdateLotTons` / `total_tons` / tentative-prices que conviven en los mismos `usecases.go`/`handler.go` (DONE en #117/#121/#124 — ver "YA PORTEADO").

## Comportamiento esperado

- `GET .../export` (lot, work-order, stock), `GET /supplies/export/all`, `GET .../export` (supply/stock movements), y las dos variantes de labor devuelven `200` con `Content-Type: text/csv; charset=utf-8`, BOM UTF-8, primera línea `sep=;`, separador `;`, fechas `02/01/2006`, decimales con `strconv.FormatFloat(..,'f',2,64)`.
- Sin filas: `404 NotFound` ("there is no data to export").
- work-order agrega fila `TOTAL` (superficie/consumo/costo sumados; las columnas dosis/cost/precio en TOTAL salen `"0.00"`).

## Estado en dp~1 (777e5f6a)

Refactor **completo y consistente** en el source: cero referencias residuales a `xuri/excelize`, `XLSXEnginePort` o `ExportToWriter` en todo el árbol; wire regenerado; tests de export actualizados a CSV. Compila y testea en el contexto de dp~1 (que ya está sobre platform/*).

## Criterios de aceptación

1. Existe `internal/shared/csvexport/writer.go` y los 5 `csv-service.go`; no existe ningún `excel-service.go`, subpaquete `excel/`, ni `internal/platform/files/excel/excelize/`.
2. `go.mod`/`go.sum` sin `github.com/xuri/excelize/v2`.
3. `grep -rn "excelize\|XLSXEnginePort\|ExportToWriter\|NewExcelExporter"` no devuelve nada.
4. `go build ./...` y `go test ./internal/lot/... ./internal/labor/... ./internal/stock/... ./internal/supply/... ./internal/work-order/... ./internal/shared/csvexport/...` en verde.
5. Los handlers de export responden `text/csv; charset=utf-8`.

## Dependencias

- **Intra-repo:** depende del paquete `domainerr` de platform/core (la firma `domainerr.NotFound/Internal`). En dp~1 es `platform/errors/go`; en `develop` confirmar el import correcto antes de copiar `writer.go` (riesgo de import path core vs platform).
- **Cross-repo:** ninguna dependencia dura. Coordinación: ver "alcance en el otro repo" (BFF FE quedaría sirviendo CSV como .xlsx).

## Riesgos (resumen — detalle en risks.md)

- **Funcional:** descarga con nombre `.xlsx` desde el BFF FE; columnas/orden distintos a los del XLSX previo (no auditado columna a columna).
- **Técnico:** `wire_gen.go`/`go.mod`/`go.sum` y `usecases.go`/`handler.go` traen hunks de OTRAS features; copiar enteros rompe o arrastra new-cns3.
- **Extracción parcial:** los `csv-service.go` no compilan solos sin los renames en usecases/handler/wire (no están en el flist).

## DECISIÓN recomendada

**Partir en subfeatures + arreglar antes (no extraer tal cual el flist).** Los 8 archivos del flist (1 writer + 5 csv-service + borrados) son el núcleo limpio y se extraen whole-file. Pero la feature NO compila sin: (a) los renames `excel->exporter` en los 5 `usecases.go` (partial-hunks), (b) el switch de Content-Type/filename en los 5 `handler.go` (partial-hunks), (c) los `wire/*_providers.go` (hunks limpios, casi whole), (d) la edición quirúrgica de `wire/wire_gen.go` (regenerar con `wire`, NO copiar), (e) borrar `internal/platform/files/excel/excelize/*` y la línea `excelize` de `go.mod`/`go.sum` (editar a mano, NO copiar enteros por la churn de platform). Mergear BE-first; abrir follow-up FE para el BFF.
