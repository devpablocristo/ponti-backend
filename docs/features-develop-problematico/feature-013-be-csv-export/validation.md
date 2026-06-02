# feature-013 · be-csv-export — validation

## Checklist pre-PR (BE)

- [ ] Existen los 6 nuevos: `internal/shared/csvexport/writer.go`, `internal/{labor,lot,stock,supply,work-order}/csv-service.go`.
- [ ] No existe ningún `excel-service.go`, subpaquete `internal/*/excel/`, ni `internal/platform/files/excel/excelize/`.
- [ ] `grep -rn "excelize\|XLSXEnginePort\|ExportToWriter\|NewExcelExporter\|PkgExcelService\|u\.excel\b" internal/ wire/` -> **0 resultados**.
- [ ] `go.mod` sin `github.com/xuri/excelize/v2`; `go.sum` sin sus 2 líneas; `go mod tidy` limpio.
- [ ] `wire/wire_gen.go` regenerado (cada dominio: `exporterAdapterPort := ProvideXExporterPort()` sin engine ni `err`).
- [ ] `internal/supply/mocks/mock_repository.go`: `MockExporterAdapterPort` solo con `ExportSupplies`/`ExportSupplyMovements`.
- [ ] El diff del PR NO contiene `UpdateLotTons`, `total_tons`, `GetTentativePrices`, ni la migración core->platform de `go.mod`.

## Comandos de build/test

```
# build completo
go build ./...

# tests de los paquetes afectados
go test ./internal/shared/csvexport/... \
        ./internal/lot/... \
        ./internal/labor/... \
        ./internal/stock/... \
        ./internal/supply/... \
        ./internal/work-order/... \
        ./wire/...

# higiene
go vet ./...
git diff --check
go mod verify
```

## Tests sugeridos a agregar (BE)

- `internal/shared/csvexport/writer_test.go`:
  - `Write(nil, nil)` / `rows` vacío -> error `NotFound` ("there is no data to export").
  - Salida empieza con BOM `0xEF 0xBB 0xBF` y luego `sep=;\n`.
  - Campos con `;`, `"` o saltos de línea quedan quoteados (RFC4180) y el separador es `;`.
- Reusar/mantener `internal/lot/handler_export_test.go` (afirma Content-Type `text/csv; charset=utf-8` y cuerpo CSV).

## Validación manual (API)

Probar cada endpoint export (con datos y sin datos):

| dominio | ruta | con datos | sin datos |
|---|---|---|---|
| lot | `GET /lots/export?project_id=...` | 200, `text/csv`, filename `lotes.csv` | 404 |
| work-order | `GET /work-orders/export` | 200, `ordenes_de_trabajos.csv`, incluye fila TOTAL | 404 |
| stock | `GET /stock/export` | 200, `stock.csv` | 404 |
| supply (tabla) | `GET /supplies/export/all` | 200, `insumos.csv` | 404 |
| supply movements | `GET /supply-movements/export` y `/stock-movements/export` | 200, `movimientos_insumos.csv` | 404 |
| labor (proyecto) | endpoint export labor | 200, `labores.csv` | 404 |
| labor (tabla) | endpoint export tabla labor | 200, `labores_tabla.csv` | 404 |

Verificar en cada respuesta: header `Content-Type: text/csv; charset=utf-8`, `Content-Disposition: attachment; filename="...csv"`, cuerpo abre limpio en Excel/Sheets (acentos OK por BOM, columnas separadas por `;`).

## Casos borde

- Export con 0 filas -> 404 (no archivo vacío).
- Decimales: `decToString` redondea a 2 (`'f',2,64`); verificar montos grandes y negativos.
- Fechas nil (sowing/harvest/invoice/movement/close) -> celda vacía, no panic.
- work-order TOTAL: superficie/consumo/costo sumados; dosis/cost-u$/precio en TOTAL = `"0.00"`.
- supply movement: `it.Investor`, `it.Provider`, `it.Supply` no nil (en stock sí se chequea nil; en supply movement NO -> revisar posible panic si vienen nil).

## Qué revisar en UI/API/DB/env

- **API:** Content-Type/filename (arriba).
- **DB:** nada (sin migraciones).
- **env:** ya NO se usa `os.TempDir()` para el archivo Excel; verificar que no quede config muerta de Excel en `cmd/config`/loadconfig (no esperado, pero confirmar con grep `excel` en config).

## Qué validar en el otro repo (FE / web)

- `web/api/src/routes/{lots,labors,workorders,stock,stock_movements,movements}.ts`: hoy fijan `filename="*.xlsx"` y `Content-Type: ...spreadsheetml.sheet`. Tras el merge BE, decidir follow-up: cambiar a `.csv`/`text/csv` o reenviar el Content-Type del upstream. Probar la descarga end-to-end desde la UI.
- Confirmar que el FE no parsea el binario como XLSX en cliente (no debería; el BFF reenvía bytes).

## Señales de incompletitud / incompatibilidad

- Build error `undefined: ProvideLotPkgExcelService` / `lot.NewExcelExporter` -> port parcial.
- `go mod tidy` reintroduce excelize -> quedó un import vivo (probable engine plataforma no borrado).
- Test `handler_export_test.go` rojo por Content-Type -> handler no migrado.
- Diff con cientos de líneas en `go.mod`/`go.sum` -> se coló new-cns3.
