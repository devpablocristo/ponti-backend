# file-list.md — feature-027 BE cleanup & domain purity

Fuente autoritativa: `/tmp/flists/be-027.txt` (2 entradas). Diff base: `0972e565..777e5f6a`.

Leyenda extracción: `whole-file` / `partial-hunks` / `manual-port` / `do-not-extract-yet`.

## Propios (núcleo de la feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/report/usecases/domain/field-crop.go` | M | dominio Go (structs) | Quita tags `json:`/`gorm:` de `ReportFilter`, `ProjectInfo`, `FieldCropMetric`, `FieldCrop`, `FieldCropColumn`, `FieldCropRow`, `FieldCropValue`, `LaborMetric`, `SupplyMetric`; reescribe comentario de `ProjectInfo` apuntando a `repository/models/project_info.go` | **whole-file** | El diff es 100% remoción de tags + 1 comentario; no se mezcla con otra intención. Traer el archivo entero del SOURCE es seguro | bajo | alta |
| `internal/report/usecases/domain/summary-results.go` | M | dominio Go (structs) | Quita tags `json:` de `SummaryResults`, `ProjectTotals`, `GeneralCrops`, `SummaryResultsResponse`, `SummaryResultsFilter` | **whole-file** | Igual que arriba; cambio mecánico y aislado | bajo | alta |

## Compartidos (partial-hunks)

Ninguno. Ambos archivos son exclusivos de esta intención de "domain purity". No tocan routers/registries/bootstrap/lockfiles ni `internal/shared/**`. No son archivos compartidos.

## Requeridos por dependencia (NO en este flist — deben existir en develop, no extraer aquí)

| path | status en rango | rol | extracción | motivo |
|---|---|---|---|---|
| `internal/report/handler/dto/field-crop.go` | (M en rango, otra feature) | Define el contrato JSON real (`json:"..."`) — es quien serializa | **do-not-extract-yet** (ya presente en develop) | Verificado PRESENTE en develop. Es prerequisito de que quitar json del dominio sea inocuo |
| `internal/report/handler/dto/summary-results.go` | (otra feature) | DTO de summary | do-not-extract-yet (presente en develop) | Prerequisito de contrato |
| `internal/report/usecases/mappers/summary_mappers.go` | (otra feature) | Mapea domain→domain/response | do-not-extract-yet (presente en develop) | Prerequisito |
| `internal/report/handler.go` | M (otra feature) | Llama `dto.BuildFieldCropResponse`, `dto.FromDomainSummaryResults` antes de `c.JSON` | do-not-extract-yet (presente en develop) | Confirma que NO se serializa el dominio directo |
| `internal/report/repository/models/project_info.go` | A (feature 001) | Shape SQL con `gorm:column:` + `ToDomain()`; es a lo que apunta el comentario nuevo | **do-not-extract-yet** (AUSENTE en develop) | Pertenece a 001-be-platform-tenancy-refactor. No bloqueante para compilar/runtime de 027 |

## Dudosos

Ninguno. El flist es inequívoco (2 archivos, ambos M, mismo patrón de cambio).

## NO traer todavía (mencionados en la NOTA de la feature pero FUERA de este flist)

| path | status en rango | por qué NO aquí |
|---|---|---|
| `internal/shared/utils/jwt_tools.go` | D | "borrar jwt utils legacy" de la nota, pero NO está en `be-027.txt`. Pertenece a feature de JWT/identity (007/008) o a la limpieza de 001. Extraer por separado |
| `internal/platform/http/middlewares/gin/require_jwt.go` | D | Idem JWT, fuera de flist |
| `internal/axis/jwt.go` | A | Reemplazo de JWT, fuera de flist |
| core/governance (sin path concreto en el flist) | — | "remove core/governance" de la nota; no hay archivo en `be-027.txt`. Probablemente ya removido o asignado a otra feature |
| staticcheck / golangci config | — | No aparece en el flist |

> NOTA de honestidad: la NOTA textual de feature-027 ("staticcheck + report domain json-tag removal + remove core/governance + borrar jwt utils legacy") describe un alcance MÁS amplio que el flist real. El flist solo cubre el bloque "report domain json-tag removal". Todo el resto se documenta como NO-traer-aquí.
