# feature-013 · be-csv-export — file-list

Fuente: `cat /tmp/flists/be-013.txt` (23 paths). Status: A=created, D=deleted.
SOURCE = `develop-problematico~1` (`777e5f6a`). Base = `develop`.

> Nota: el flist contiene SOLO el núcleo (writer + 5 csv-service nuevos + borrados Excel por dominio). Los archivos que CONSUMEN estos exporters (handlers, usecases, wire, go.mod/go.sum, mocks, el engine excelize de plataforma) NO están en el flist pero son imprescindibles para compilar. Los listo en "Compartidos" y "Requeridos por dependencia".

Leyenda extracción: `whole-file` | `partial-hunks` | `manual-port` | `do-not-extract-yet`.

## Propios (núcleo de la feature — del flist)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/shared/csvexport/writer.go` | A | go pkg compartido | API central `Write(headers, rows)` (BOM, `sep=;`, RFC4180) | whole-file | archivo nuevo, autocontenido | bajo — solo verificar import `domainerr` (core vs platform) | alta |
| `internal/labor/csv-service.go` | A | go service | `CSVExporter.Export`/`ExportTable` labor | whole-file | nuevo, depende solo de `domain` + `csvexport` + `decimal` | bajo | alta |
| `internal/lot/csv-service.go` | A | go service | `CSVExporter.Export([]domain.LotTable)` | whole-file | nuevo, autocontenido | bajo | alta |
| `internal/stock/csv-service.go` | A | go service | `CSVExporter.Export([]*domain.Stock)` | whole-file | nuevo, usa getters de `domain.Stock` | bajo | alta |
| `internal/supply/csv-service.go` | A | go service | `CSVExporter.ExportSupplies`/`ExportSupplyMovements` | whole-file | nuevo | bajo | alta |
| `internal/work-order/csv-service.go` | A | go service | `CSVExporter.Export` + fila TOTAL | whole-file | nuevo | bajo | alta |
| `internal/labor/excel-service.go` | D | borrado | reemplazado por csv-service | whole-file (borrar) | el ExporterAdapterPort ahora lo da CSV | bajo | alta |
| `internal/labor/excel/config.go` | D | borrado | config Excel | whole-file (borrar) | — | bajo | alta |
| `internal/labor/excel/excel-dto.go` | D | borrado | DTO Excel proyecto | whole-file (borrar) | — | bajo | alta |
| `internal/labor/excel/excel-table-dto.go` | D | borrado | DTO Excel tabla | whole-file (borrar) | — | bajo | alta |
| `internal/lot/excel-service.go` | D | borrado | reemplazado | whole-file (borrar) | — | bajo | alta |
| `internal/lot/excel/config.go` | D | borrado | config Excel lot | whole-file (borrar) | — | bajo | alta |
| `internal/lot/excel/excel-dto.go` | D | borrado | DTO Excel lot | whole-file (borrar) | — | bajo | alta |
| `internal/stock/excel-service.go` | D | borrado | reemplazado | whole-file (borrar) | — | bajo | alta |
| `internal/stock/excel/config.go` | D | borrado | config Excel stock | whole-file (borrar) | — | bajo | alta |
| `internal/stock/excel/excel-dto.go` | D | borrado | DTO Excel stock | whole-file (borrar) | — | bajo | alta |
| `internal/supply/excel-service.go` | D | borrado | reemplazado | whole-file (borrar) | — | bajo | alta |
| `internal/supply/excel/config.go` | D | borrado | config Excel supply | whole-file (borrar) | — | bajo | alta |
| `internal/supply/excel/excel_dto.go` | D | borrado | DTO Excel supply | whole-file (borrar) | — | bajo | alta |
| `internal/supply/excel/excel_dto_table.go` | D | borrado | DTO Excel supply tabla | whole-file (borrar) | — | bajo | alta |
| `internal/supply/excel/helpers.go` | D | borrado | helpers Excel supply | whole-file (borrar) | — | bajo | alta |
| `internal/work-order/excel-service.go` | D | borrado | reemplazado | whole-file (borrar) | — | bajo | alta |
| `internal/work-order/excel/config.go` | D | borrado | config Excel WO | whole-file (borrar) | — | bajo | alta |
| `internal/work-order/excel/excel-dto.go` | D | borrado | DTO Excel WO | whole-file (borrar) | — | bajo | alta |

## Compartidos (partial-hunks — NO en el flist, hunks mezclados con otras features)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/lot/usecases.go` | M | go | rename `excel`->`exporter`, `u.excel.Export`->`u.exporter.Export` | partial-hunks | mezcla con lot-metrics `UpdateLotTons`/`total_tons` (DONE) | alto | media |
| `internal/labor/usecases.go` | M | go | rename + `ExportTable` en interfaz | partial-hunks | conviven cambios de tenancy/dominio | alto | media |
| `internal/stock/usecases.go` | M | go | `u.exporter.Export` | partial-hunks | mezcla con stock filter/dominio | alto | media |
| `internal/supply/usecases.go` + `usecases_movement.go` | M | go | `ExportSupplies`/`ExportSupplyMovements` en interfaz + calls | partial-hunks | mezcla con movimientos/import | alto | media |
| `internal/work-order/usecases.go` | M | go | `u.exporter.Export` | partial-hunks | mezcla con cambios WO | alto | media |
| `internal/lot/handler.go` | M | go | import csvexport, Content-Type/filename, `c.Data` | partial-hunks | mezcla con `UpdateLotTons`, `RespondNoContent` (DONE/otras) | alto | media |
| `internal/labor/handler.go` | M | go | 3 endpoints export -> CSV headers | partial-hunks | idem | alto | media |
| `internal/stock/handler.go` | M | go | export -> CSV | partial-hunks | idem | alto | media |
| `internal/supply/handler.go` | M | go | `ExportTableSupplies`, `ExportSupplyMovementsByProjectID` -> CSV | partial-hunks | rutas `/export/all`, `/export` | alto | media |
| `internal/work-order/handler.go` | M | go | export `/export` -> CSV | partial-hunks | idem | alto | media |
| `wire/lot_providers.go` | M | go wire | quita Excel providers, `ProvideLotExporterPort()` CSV | partial-hunks (casi whole) | hunk dedicado y limpio a 013 | medio | alta |
| `wire/labor_providers.go` | M | go wire | idem labor | partial-hunks (casi whole) | hunk limpio a 013 | medio | alta |
| `wire/stock_providers.go` | M | go wire | idem stock | partial-hunks (casi whole) | hunk limpio a 013 | medio | alta |
| `wire/supply_providers.go` | M | go wire | idem supply | partial-hunks (casi whole) | hunk limpio a 013 | medio | alta |
| `wire/work_order_providers.go` | M | go wire | idem WO (`ProvideWorkOrderExporterPort()`) | partial-hunks (casi whole) | hunk limpio a 013 | medio | alta |
| `wire/wire.go` | M | go wire | quita providers Excel del set/registro | partial-hunks | mezcla con todas las features | alto | media |
| `wire/wire_gen.go` | M | go generado | quita líneas `...ExcelService`/`XLSXEnginePort` | manual-port (regenerar) | NUNCA copiar: mezcla todas las features; correr `wire` | muy alto | media |
| `go.mod` | M | deps | quitar `github.com/xuri/excelize/v2 v2.9.1` | manual-port | resto del diff es new-cns3 (core->platform) — NO traer | muy alto | alta |
| `go.sum` | M | deps | quitar 2 líneas `xuri/excelize/v2` | manual-port | resto es churn platform | muy alto | alta |
| `internal/supply/mocks/mock_repository.go` | M | go mock | `MockExporterAdapterPort` con `ExportSupplies`/`ExportSupplyMovements` | manual-port (regenerar mocks) | mock generado; mejor regenerar | medio | media |

## Requeridos por dependencia (NO en el flist, parte de 013)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/platform/files/excel/excelize/bootstrap.go` | D | borrado | engine Excel de plataforma | whole-file (borrar) | sin consumidores tras quitar wire Excel | medio | alta |
| `internal/platform/files/excel/excelize/config.go` | D | borrado | config engine | whole-file (borrar) | idem | medio | alta |
| `internal/platform/files/excel/excelize/service.go` | D | borrado | `pkgexcel.Service`/`Bootstrap` | whole-file (borrar) | idem | medio | alta |

## Tests (NO en el flist, deben acompañar)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/lot/handler_export_test.go` | M | test | stub devuelve `"csv"`, Content-Type `text/csv` | partial-hunks | valida el contrato CSV | medio | alta |
| `internal/labor/handler_update_labor_test.go` | M | test | stub de UseCases con `ExportTable` | partial-hunks | compila los tests | medio | media |
| `internal/supply/handler_update_supply_test.go` | M | test | stub `ExportTableSupplies`/`ExportSupplyMovementsByProjectID` | partial-hunks | compila los tests | medio | media |
| `internal/work-order/handler_test.go` | M | test | ajuste de export | partial-hunks | compila | medio | media |

## NO traer todavía (excluir — ya DONE u otra feature)

| path/area | motivo |
|---|---|
| Hunks `UpdateLotTons` / `total_tons` en `lot/usecases.go`,`lot/handler.go` | lot-metrics DONE (#117/#121/#124) — excluir esos hunks |
| Hunks `GetTentativePrices`/tentative-prices | DONE (#121/#124) — excluir |
| Hunks `RespondNoContent`/`shared/handlers/*` | pertenecen a otras features (008/018) salvo lo mínimo que toque export |
| Todo el diff core->platform en `go.mod`/`go.sum`/imports | new-cns3 (no 013) — solo tomar la línea excelize |
| `internal/*/repository_*_test.go` nuevos (tenant/archived/crudar) | tests de 001/002/003/009 — NO 013 |
| `internal/shared/{authz,filters,lifecycle,text}/*_test.go` | otras features |
