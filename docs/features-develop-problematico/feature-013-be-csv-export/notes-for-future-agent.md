# feature-013 · be-csv-export — notes for future agent

## Resumen corto

Refactor solo-BE: reemplaza el export XLSX (`xuri/excelize`) por CSV. Núcleo = paquete compartido `internal/shared/csvexport` (`Write(headers, rows)`: BOM UTF-8, `sep=;`, separador `;`, RFC4180) + un `csv-service.go` por dominio (labor, lot, stock, supply, work-order). Borra el engine Excel de plataforma y la dep excelize. El refactor en el SOURCE está 100% completo y limpio.

## Qué está en FE y qué en BE

- **BE (este repo):** toda la lógica. Núcleo del flist + consumidores (handlers/usecases/wire) + borrados.
- **FE (web):** NO hay feature. PERO el BFF `web/api/src/routes/{lots,labors,workorders,stock,stock_movements,movements}.ts` hardcodea `filename="*.xlsx"` y MIME spreadsheet. Es un follow-up FE, NO parte de 013.

## Archivos esenciales

- `internal/shared/csvexport/writer.go` — el corazón. Reusar import `domainerr` correcto (core vs platform según `develop`).
- Los 5 `internal/*/csv-service.go` — definen headers y mapeo de columnas.
- `wire/*_providers.go` — wiring CSV (hunks limpios, casi whole-file).

## Archivos peligrosos / mezclados (NO copiar enteros)

- `wire/wire_gen.go` — REGENERAR (no copiar literal).
- `go.mod` / `go.sum` — solo quitar `xuri/excelize/v2`; el resto del diff es new-cns3 (core->platform).
- `internal/*/usecases.go` y `internal/*/handler.go` — partial-hunks; conviven con lot-metrics (DONE), tentative-prices (DONE), tenancy, lifecycle.
- `internal/supply/mocks/mock_repository.go` — regenerar.
- `internal/platform/files/excel/excelize/*` — BORRAR (no está en el flist pero es parte de 013).

## Decisiones ya tomadas

- DECISIÓN: extraer núcleo whole-file + partial-hunks para consumidores + regenerar wire_gen/mocks + editar go.mod a mano. NO extraer el flist tal cual (no compila solo).
- Orden: **BE-first**; FE (BFF) es follow-up opcional.
- Sin migraciones, sin cambios de DB.

## Dudas abiertas

- Import path de `domainerr` en `develop` (core vs platform).
- ¿Hay tooling `wire`/`mockgen` en el repo para regenerar? (revisar Makefile).
- ¿Se ajusta el BFF FE en el mismo ciclo o después? (decisión humana).

## Comandos para mirar primero

```
cat /tmp/flists/be-013.txt
git -C <core> grep -rn "excelize\|XLSXEnginePort\|NewCSVExporter\|ProvideX.*ExporterPort" 777e5f6a
git -C <core> show 777e5f6a:internal/shared/csvexport/writer.go
git -C <core> diff 0972e565..777e5f6a -- wire/lot_providers.go        # hunk limpio de ejemplo
git -C <core> diff 0972e565..777e5f6a -- internal/lot/handler.go | grep -iE "excel|csv|content"  # ver mezcla
git -C <core> grep -n "domainerr" develop -- internal/shared           # resolver import en destino
```

## Errores a evitar

- NO usar `develop-problematico` (tip vacío); usar `develop-problematico~1` = `777e5f6a`.
- NO copiar `go.mod`/`go.sum`/`wire_gen.go` enteros.
- NO arrastrar hunks de `UpdateLotTons`/`total_tons`/`GetTentativePrices` (ya porteados).
- NO olvidar borrar el engine `internal/platform/files/excel/excelize/` (no está en el flist).
- NO mergear FE antes que BE.

## Camino más seguro

1. Branch desde `develop`.
2. `git checkout 777e5f6a --` los 6 nuevos + los 5 `wire/*_providers.go`.
3. `git rm -r` Excel por dominio + engine excelize.
4. `git restore -p` los hunks export de usecases/handler/wire.go/tests.
5. Editar `go.mod`/`go.sum` (quitar excelize) + `go mod tidy`.
6. Regenerar `wire_gen.go` y mocks.
7. `grep` de residuales = 0, `go build`, `go test`, PR.

## PR del otro repo (FE) antes/después

- **Después** del merge BE: PR FE en `web` para que el BFF de export devuelva `.csv`/`text/csv`. Pertenece al área FE (014/017) o un fix de BFF aislado. No bloquea ni precede a 013.
