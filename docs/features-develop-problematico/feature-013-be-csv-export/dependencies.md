# feature-013 · be-csv-export — dependencies

## Depende-de

| dep | fuerza | tipo | detalle |
|---|---|---|---|
| paquete `domainerr` (platform/core errors) | **fuerte** | intra-repo | `csvexport.Write` usa `domainerr.NotFound`/`domainerr.Internal`. En el source (`777e5f6a`) el import es `github.com/devpablocristo/platform/errors/go/domainerr`. En `develop` confirmar el path antes de copiar `writer.go`. |
| `github.com/shopspring/decimal` | fuerte | intra-repo (dep ya presente) | helper `decToString` en cada csv-service. Ya está en `go.mod` de `develop`. |
| `encoding/csv`, `bytes` (stdlib) | fuerte | stdlib | sin riesgo. |
| dominios `internal/*/usecases/domain` | fuerte | intra-repo | cada csv-service lee campos de los DTOs de dominio (`LotTable`, `Stock`, `Supply`, `SupplyMovement`, `WorkOrderListElement`, `LaborListItem`, `ListedLabor`). Esos tipos deben existir igual en `develop`. |
| ninguna feature 001..027 | — | — | 013 NO depende de otra feature para su lógica. |

## Bloquea-a

| bloquea | fuerza | detalle |
|---|---|---|
| follow-up FE del BFF de export | **débil** | El cambio de Content-Type/filename a CSV en BE invita (no obliga) a actualizar `web/api/src/routes/*.ts`. No bloquea funcionalmente (Excel abre el CSV). |
| nada más | — | — |

## Inciertas

| incertidumbre | qué revisar | confianza |
|---|---|---|
| import path `domainerr` core vs platform en `develop` | `git -C core grep -n "errors/go/domainerr" -- internal/shared` en `develop`; comparar con writer.go del source | media |
| si `develop` ya tiene los renames `excel->exporter` parciales (por algún cherry-pick previo) | `git -C core grep -rn "u.excel\b\|exporter ExporterAdapterPort" develop -- internal/` | media |
| estado de `wire_gen.go` en `develop` (¿generado con qué versión de wire?) | leer `wire/wire_gen.go` en `develop` y Makefile target `wire`/`generate` | media |

## Compartido (archivos/tipos/config/APIs que tocan a 013 y a otras features)

- **`wire/wire_gen.go` + `wire/wire.go`** — registro DI global; mezcla TODAS las features. Compartido con 023 (be-wire-di) y todas las que agregan providers. Editar quirúrgico / regenerar.
- **`go.mod` / `go.sum`** — 013 solo quita `xuri/excelize/v2`. El grueso del diff en el rango es la migración core->platform (021/new-cns3). Compartido con 021 (build/deploy) y 005 (config) en cuanto a deps.
- **`internal/*/usecases.go` y `internal/*/handler.go`** — comparten archivo con lot-metrics (DONE), tentative-prices (DONE), tenancy (001/003), lifecycle/crudar (002/009). Por eso son partial-hunks.
- **`internal/supply/mocks/mock_repository.go`** — mock generado compartido con tests de supply (025).
- **APIs (rutas)** — las rutas `/export*` no cambian de path; cambia el contrato de respuesta (Content-Type/filename). Contrato consumido por el BFF FE (cross-repo).

## Cross-repo

- **FE (web):** sin carpeta de feature. Consumo del contrato vía BFF `web/api/src/routes/{lots,labors,workorders,stock,stock_movements,movements}.ts`, que hoy hardcodean `.xlsx` + MIME spreadsheet. Dependencia **débil/diferida**: el BFF debería pasar a `.csv`/`text/csv` en un PR FE posterior. No hay tipos compartidos formales (BFF reenvía bytes opacos).

## Recomendación de orden

1. **BE-first:** mergear feature-013 en `develop` del backend (esta extracción).
2. **FE después (opcional, recomendado):** PR en `web` ajustando el BFF de export a CSV. Encaja como follow-up del área 014/017 o un fix de BFF independiente; NO es parte de 013.
3. No requiere coordinación de release simultánea: BE puede ir solo sin romper FE (el archivo descargado abre igual; solo queda mal el nombre/MIME hasta el fix FE).
