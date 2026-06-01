# risks.md — feature-027 BE cleanup & domain purity

## Riesgos funcionales

| riesgo | prob | impacto | mitigación |
|---|---|---|---|
| Algún punto del código serializa `domain.*` de report directamente (sin pasar por DTO), por lo que quitar `json:` cambiaría los nombres de campo en esa respuesta | baja | alto (rompe contrato del FE en ese endpoint) | `grep -rn 'c.JSON' internal/report/` y confirmar que siempre se pasa un DTO. Verificado en `handler.go`: usa `dto.BuildFieldCropResponse` / `dto.FromDomainSummaryResults` / `dto.FromDomainInvestorReport` antes de `c.JSON`. Riesgo residual: otros handlers que importen estos tipos |
| Respuesta JSON con nombres en CamelCase en vez de snake_case | baja | alto | Comparar la respuesta del endpoint de report contra develop pre-cambio (diff de payload) |

## Riesgos técnicos

| riesgo | prob | impacto | mitigación |
|---|---|---|---|
| `getProjectInfo` deja de mapear columnas porque se quitaron los `gorm:"column:..."` de `domain.ProjectInfo` | baja | medio (campos en cero/vacío) | El query aliasea columnas en snake_case (`project_id`, `customer_name`, ...) y GORM default snake-casea los campos → mapea igual. Confirmar NamingStrategy de GORM si se quiere certeza. Smoke del endpoint que usa ProjectInfo |
| No compila por algún consumidor que esperaba tags (no aplica: tags no son parte de la firma de tipo) | muy baja | medio | `go build ./...` |

## Riesgos de integración

- Ninguno cross-servicio. Cambio interno al módulo report.

## Riesgos cross-repo

- Ninguno. Solo-BE. El FE no consume estos structs directamente; consume el JSON, que NO cambia.

## Riesgos de datos / migración

- Ninguno. Sin migraciones, sin cambios de esquema, sin cambios de datos.

## Riesgos de archivos compartidos

- Ninguno. Los 2 archivos NO son compartidos: no tocan `wire/`, `cmd/`, `go.mod`, `go.sum`, `Makefile`, `internal/shared/**`. No hay hunks que sirvan a varias intenciones.

## Riesgos de extracción parcial

| riesgo | mitigación |
|---|---|
| Traer solo uno de los 2 archivos | El flist tiene exactamente 2; el plan trae ambos. Verificar `git diff --staged --name-only` == 2 paths |
| Que queden tags residuales (limpieza incompleta) | `grep -rn 'json:"\|gorm:"' internal/report/usecases/domain/field-crop.go internal/report/usecases/domain/summary-results.go` debe dar 0 |
| Arrastrar por error tests o `repository/models/project_info.go` que pertenecen a otras features | El comando `git checkout dp~1 -- <2 paths>` solo trae esos 2 paths; no usar globs amplios |

## Riesgo de mergear SOLO este repo

- **Bajo.** Es BE independiente y no requiere coordinación FE (no hay cambios FE). El contrato JSON no cambia, así que el FE no nota nada. Mergear solo BE es seguro.

## Riesgo de mergear SOLO el otro repo

- N/A: no hay cambios FE para esta feature.

## Resumen de severidad

- Feature de **bajo riesgo global**. Cambio mecánico, aislado, con la dependencia fuerte (capa DTO) ya satisfecha en develop. Los riesgos restantes son improbables y detectables con build + un smoke de endpoint.
