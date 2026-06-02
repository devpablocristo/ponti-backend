# feature-013 · be-csv-export — extraction-plan

- **repo:** ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip `003a9b8f`)
- **SOURCE:** `develop-problematico~1` = SHA `777e5f6a` (merge PR #120). NUNCA `develop-problematico` (tip restore/vacío).
- **rango de verdad:** `0972e565..777e5f6a`
- **rama sugerida:** `pr/feature-013-be-csv-export-be`
- **orden cross-repo:** **BE-first.** El FE no tiene feature; el único follow-up FE (BFF) puede ir después y es opcional para que la descarga funcione (Excel abre el CSV igual).

## PR title

`refactor(be): reemplazar export XLSX por CSV (csvexport + csv-service por dominio)`

## PR description (sugerida)

> Reemplaza el motor de exportación XLSX (`xuri/excelize`) por un exportador CSV compartido.
>
> - Nuevo `internal/shared/csvexport.Write` (UTF-8 BOM, `sep=;`, `;`, RFC4180).
> - `csv-service.go` por dominio (labor, lot, stock, supply, work-order) reemplaza `excel-service.go` + subpaquete `excel/`.
> - Borra el engine `internal/platform/files/excel/excelize/*` y la dep `xuri/excelize/v2`.
> - Wire simplificado: `ProvideXExporterPort()` sin engine; regenerado `wire_gen.go`.
> - Handlers de export ahora responden `text/csv; charset=utf-8` con filename `.csv`.
>
> NO incluye: cambios en BFF FE (follow-up), ni la migración core->platform (es de new-cns3).
>
> Endpoints afectados (sin cambio de ruta, sí de Content-Type/filename): `GET /lots/export`, `GET /work-orders/export`, `GET /stock/export`, `GET /supplies/export/all`, `GET /supply-movements/export`, `GET /stock-movements/export`, y los 2 export de labor.

## Pasos ordenados

1. `git checkout develop && git pull`
2. `git checkout -b pr/feature-013-be-csv-export-be`
3. **Whole-file nuevos** (núcleo, del flist):
   `git checkout 777e5f6a -- internal/shared/csvexport/writer.go internal/labor/csv-service.go internal/lot/csv-service.go internal/stock/csv-service.go internal/supply/csv-service.go internal/work-order/csv-service.go`
4. **Borrados** (del flist) — borrar los `excel-service.go` y subpaquetes `excel/` de los 5 dominios:
   `git rm -r internal/labor/excel-service.go internal/labor/excel internal/lot/excel-service.go internal/lot/excel internal/stock/excel-service.go internal/stock/excel internal/supply/excel-service.go internal/supply/excel internal/work-order/excel-service.go internal/work-order/excel`
5. **Borrado del engine de plataforma** (requerido por dependencia, NO en flist):
   `git rm -r internal/platform/files/excel/excelize`
6. **Wire providers** (hunks limpios, casi whole-file) — traerlos del source:
   `git checkout 777e5f6a -- wire/lot_providers.go wire/labor_providers.go wire/stock_providers.go wire/supply_providers.go wire/work_order_providers.go`
   - REVISAR cada uno: que NO arrastren imports core->platform que en `develop` aún no existan. Si arrastran, editar a mano (solo el bloque Excel->CSV).
7. **usecases.go (partial)** — por dominio, traer SOLO los hunks de rename `excel`->`exporter` y `u.exporter.Export`:
   `git restore -p --source=777e5f6a -- internal/lot/usecases.go internal/labor/usecases.go internal/stock/usecases.go internal/supply/usecases.go internal/supply/usecases_movement.go internal/work-order/usecases.go`
   - **RECHAZAR** hunks de `UpdateLotTons`/`total_tons`/tentative-prices (DONE) y de tenancy/dominio.
8. **handler.go (partial)** — traer SOLO los hunks de export (import `csvexport`, Content-Type, filename, `c.Data`):
   `git restore -p --source=777e5f6a -- internal/lot/handler.go internal/labor/handler.go internal/stock/handler.go internal/supply/handler.go internal/work-order/handler.go`
   - **RECHAZAR** hunks de `RespondNoContent`/rutas no-export/otras features.
9. **wire.go (partial)** — quitar del set/registro los providers Excel:
   `git restore -p --source=777e5f6a -- wire/wire.go` (aceptar solo hunks que borran `ProvideX{PkgExcelService,XLSXEnginePort}`).
10. **wire_gen.go** — NO copiar. Regenerar: `go generate ./wire/...` o `wire ./wire/...` (según herramienta del repo; ver Makefile). Si no hay tooling, editar a mano quitando las 3 líneas `...ExcelService`/`XLSXEnginePort`/`...ExporterPort(eng)` por dominio y dejando `exporterAdapterPort := ProvideXExporterPort()`.
11. **go.mod / go.sum** — editar a MANO: borrar `github.com/xuri/excelize/v2 v2.9.1` de require en `go.mod` y las 2 líneas `xuri/excelize/v2` de `go.sum`. Luego `go mod tidy`. NO copiar los archivos enteros del source (arrastran new-cns3).
12. **mocks** — `internal/supply/mocks/mock_repository.go`: regenerar con `go generate` / `mockgen` para que `MockExporterAdapterPort` quede solo con `ExportSupplies`/`ExportSupplyMovements`. Alternativa: `git restore -p` los hunks.
13. **tests** — traer hunks CSV: `git restore -p --source=777e5f6a -- internal/lot/handler_export_test.go internal/labor/handler_update_labor_test.go internal/supply/handler_update_supply_test.go internal/work-order/handler_test.go`
14. `git diff --check` (whitespace), `go build ./...`, `go test ./...` (al menos los paquetes afectados).
15. Commit y PR contra `develop`.

## Archivos enteros vs parciales

- **Enteros (checkout):** los 6 nuevos del núcleo + los 5 `wire/*_providers.go` (revisando imports).
- **Borrados (git rm):** Excel por dominio + engine excelize.
- **Parciales (restore -p):** 5 `usecases.go`(+movement), 5 `handler.go`, `wire/wire.go`, tests.
- **Manuales:** `wire_gen.go` (regenerar), `go.mod`/`go.sum` (1 dep), mocks (regenerar).

## Migraciones / tests a incluir

- **Migraciones:** ninguna (refactor de capa de presentación/export, sin DB).
- **Tests:** los de export listados; añadir uno mínimo para `csvexport.Write` (BOM, `sep=;`, NotFound sin filas) si no existe.

## Dependencias previas

- Ninguna feature bloquea a 013. Pero `writer.go` importa `domainerr`: confirmar el import path en `develop` (core vs platform). Si `develop` ya migró a platform, el source ya lo trae como `platform/errors/go/domainerr` — verificar.

## Comandos git SUGERIDOS (solo lectura + checkout/restore; el humano decide)

```
git checkout develop
git checkout -b pr/feature-013-be-csv-export-be
git checkout 777e5f6a -- internal/shared/csvexport/writer.go internal/labor/csv-service.go internal/lot/csv-service.go internal/stock/csv-service.go internal/supply/csv-service.go internal/work-order/csv-service.go
git rm -r internal/labor/excel internal/lot/excel internal/stock/excel internal/supply/excel internal/work-order/excel internal/platform/files/excel/excelize
git rm internal/labor/excel-service.go internal/lot/excel-service.go internal/stock/excel-service.go internal/supply/excel-service.go internal/work-order/excel-service.go
git checkout 777e5f6a -- wire/lot_providers.go wire/labor_providers.go wire/stock_providers.go wire/supply_providers.go wire/work_order_providers.go
git restore -p --source=777e5f6a -- internal/lot/usecases.go internal/labor/usecases.go internal/stock/usecases.go internal/supply/usecases.go internal/supply/usecases_movement.go internal/work-order/usecases.go
git restore -p --source=777e5f6a -- internal/lot/handler.go internal/labor/handler.go internal/stock/handler.go internal/supply/handler.go internal/work-order/handler.go
git restore -p --source=777e5f6a -- wire/wire.go
git restore -p --source=777e5f6a -- internal/lot/handler_export_test.go internal/labor/handler_update_labor_test.go internal/supply/handler_update_supply_test.go internal/work-order/handler_test.go
# luego: editar go.mod/go.sum (quitar excelize) + go mod tidy + regenerar wire_gen y mocks
git diff --check
```

## Qué NO traer

- Diff core->platform de `go.mod`/`go.sum`/imports (new-cns3).
- Hunks de `UpdateLotTons`/`total_tons`/tentative-prices (DONE).
- Tests nuevos de tenancy/archived/crudar/lifecycle/authz/propername.
- `wire_gen.go` literal del source (regenerar).

## Qué podría romperse

- Si traés `usecases.go`/`handler.go` enteros: arrastrás new-cns3 + lot-metrics y rompés `develop`.
- Si copiás `go.mod`/`go.sum` enteros: cambiás todo el árbol de deps.
- Si copiás `wire_gen.go` literal: referencias a símbolos platform que en `develop` no existan -> no compila.
- Si olvidás borrar el engine `excelize` o sus consumidores en wire: import ciclo/símbolo inexistente.

## Cómo detectar extracción incompleta

- `grep -rn "excelize\|XLSXEnginePort\|ExportToWriter\|NewExcelExporter\|PkgExcelService\|\.excel\b" internal/ wire/` debe dar **0**.
- `go build ./...` falla con "undefined: ProvideLotPkgExcelService" -> quedó un consumidor sin migrar.
- `go vet`/compilador: campo `excel` vs `exporter` en algún `usecases.go`.

## Qué validar antes del PR

- `go build ./...`, `go test ./internal/... ./wire/...`.
- Handlers de export devuelven `text/csv; charset=utf-8` (test `handler_export_test.go`).
- `go.sum` consistente tras `go mod tidy` (sin excelize, sin faltantes).

## Qué hacer después de mergear

- Abrir issue/PR FE para el BFF (`web/api/src/routes/*.ts`): cambiar `filename="*.xlsx"` y Content-Type a `.csv`/`text/csv` (o pasar el Content-Type del upstream). Ver cross-repo-map de FE.
- Avisar a QA: la descarga ahora es CSV (BOM/`;`), validar apertura en Excel ES.
