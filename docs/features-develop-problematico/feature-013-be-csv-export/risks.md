# feature-013 · be-csv-export — risks

## Funcionales

| riesgo | impacto | mitigación |
|---|---|---|
| El BFF FE re-envuelve el CSV como `.xlsx` (filename + MIME spreadsheet hardcodeados en `web/api/src/routes/*.ts`). | Usuario descarga `lots.xlsx` que es CSV. Excel ES lo abre por BOM+`sep=;`, pero el nombre/MIME confunden y otros lectores pueden fallar. | Follow-up FE: cambiar filename a `.csv` y Content-Type a `text/csv` (o reenviar el upstream). Documentar a QA. No bloquea el merge BE. |
| Columnas/orden/labels del CSV difieren del XLSX previo. | Reportes con layout distinto; usuarios acostumbrados al XLSX. | Es un cambio intencional del refactor; validar con negocio/QA. Los headers están definidos explícitamente en cada `csv-service.go`. |
| Separador `;` + locale. | En locales no-ES Excel puede no respetar `;`. | El writer emite `sep=;` como primera línea (hint de Excel). Aceptable. |
| Export sin datos devuelve `404`. | El FE/BFF debe manejar 404 como "nada para exportar". | Confirmar manejo en BFF; comportamiento ya existía en el camino Excel (`NotFound`). |

## Técnicos

| riesgo | impacto | mitigación |
|---|---|---|
| Copiar `usecases.go`/`handler.go` enteros arrastra new-cns3 + lot-metrics + tentative-prices. | Rompe `develop` o reintroduce features ya porteadas. | Usar `git restore -p` y aceptar SOLO los hunks export (rename `excel->exporter`, Content-Type/filename, `c.Data`). |
| Copiar `go.mod`/`go.sum` enteros. | Cambia todo el árbol de deps a platform/* (no es 013). | Editar a mano: quitar `xuri/excelize/v2` + `go mod tidy`. |
| Copiar `wire_gen.go` literal. | Referencias a símbolos platform inexistentes en `develop` -> no compila. | Regenerar con `wire`/`go generate`. Si no hay tooling, editar las líneas Excel a mano. |
| Import `domainerr` distinto en `develop` (core vs platform). | `writer.go` no compila. | Verificar el path en `develop` y ajustar el import del archivo copiado. |
| Olvidar borrar `internal/platform/files/excel/excelize/*` (no está en el flist). | Paquete huérfano que aún importa excelize -> `go.mod` no puede quitar la dep. | Incluir el `git rm -r` del engine en el plan (paso 5). |
| Mock `mock_repository.go` con métodos XLSX viejos. | Tests de supply no compilan. | Regenerar mocks o `restore -p` los hunks. |

## Integración

| riesgo | impacto | mitigación |
|---|---|---|
| Wire: dejar un `ProvideXPkgExcelService` colgado en el set. | "undefined" en build. | `grep "PkgExcelService\|XLSXEnginePort"` debe dar 0 tras portar. |
| Tests de handler esperan `text/csv` pero handler quedó en XLSX (port parcial). | Test rojo. | Portar handler + test juntos por dominio. |

## Cross-repo

| riesgo | impacto | mitigación |
|---|---|---|
| Mergear solo BE: BFF FE sigue diciendo `.xlsx`. | Inconsistencia cosmética (nombre/MIME), no funcional. | Aceptable a corto plazo; abrir issue FE. |
| Mergear solo FE (hipotético, sin BE): BFF esperaría CSV pero BE da XLSX. | El BFF reetiquetaría XLSX como CSV o viceversa. | NO mergear FE antes que BE. Orden BE-first es obligatorio. |

## Datos / migración

- **Sin riesgo de datos ni migraciones:** la feature es de capa de export (presentación), no toca DB ni esquemas.

## Archivos compartidos (resumen de peligro)

- `wire/wire_gen.go`, `wire/wire.go`, `go.mod`, `go.sum`: peligro **alto/muy alto** — mezclan todas las features. Tratamiento manual/regeneración.
- `internal/*/usecases.go`, `internal/*/handler.go`: peligro **alto** — partial-hunks.
- `internal/supply/mocks/mock_repository.go`: peligro **medio** — regenerar.

## Extracción parcial

- **Señal de incompletitud:** build falla con `undefined: ProvideLotPkgExcelService`, `lot.NewExcelExporter`, o campo `u.excel`.
- **Señal de sobre-extracción:** aparecen en el diff `UpdateLotTons`, `total_tons`, `GetTentativePrices`, o cambios masivos en `go.mod` (platform) -> revertir esos hunks.
- **Verificación final:** `grep -rn "excelize\|XLSXEnginePort\|ExportToWriter\|NewExcelExporter\|PkgExcelService" internal/ wire/ go.mod` = 0 y `go build ./...` verde.

## Riesgo de mergear solo este repo / solo el otro

- **Solo BE (recomendado):** seguro funcionalmente; queda deuda cosmética en el nombre/MIME del archivo descargado vía BFF. Bajo riesgo.
- **Solo FE:** no aplica (no hay feature FE); el único cambio FE posible (BFF) NO debe ir antes que BE.
