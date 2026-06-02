# risks.md — feature-025 · BE test coverage sweep

## Riesgos de orden / dependencia (los más importantes)

| riesgo | severidad | detalle | mitigación |
|---|---|---|---|
| Mergear 025 antes que 001/002/009 | **ALTA** | El build de CI falla: `undefined: assertCustomerReferencesActive`, `undefined: HardDeleteWorkOrder`, `unknown field ucs in struct literal`, `GetArchivedParameters` no existe. Verificado: estos símbolos NO están en develop tip (003a9b8f). | NO abrir/mergear este PR hasta confirmar 001/002/003/009 en `develop`. Idealmente 3 sub-PRs enganchados a cada feature productora. |
| Firmas de producción divergen entre SOURCE y lo mergeado en develop | media | Si 001/002/009 aterrizaron con otra firma (orden de params, nombre de método), el test no compila. | Antes del PR, comparar firmas: `git grep -n "func assert.*ReferencesActive" develop` y `git grep -n "func.*Routes\|func NewHandler" develop -- internal/work-order/`. Ajustar el TEST (no producción). |

## Riesgos funcionales

| riesgo | severidad | mitigación |
|---|---|---|
| Falsos positivos (test verde pero producción mal) por usar stubs de usecases en handler_test | baja | Los handler tests prueban wiring de rutas/actor, no la lógica de negocio; complementados por repo tests. Aceptable. |
| sqlite vs Postgres divergen en semántica de `deleted_at` / tipos | baja | Tests miden lógica de filtrado por tenant y chequeo de archivado, no SQL Postgres-specific. Aceptable como unit test. |

## Riesgos técnicos

| riesgo | severidad | mitigación |
|---|---|---|
| Tests white-box (`package <modulo>`) acoplados a internals | media | Inevitable para `assert*ReferencesActive` (no exportados). Documentado; cualquier refactor de esas funciones debe actualizar el test. |
| `t.Setenv("TENANT_STRICT_MODE", ...)` afecta orden de tests | baja | `t.Setenv` se revierte por test (Go ≥1.17) y marca el test como no paralelizable automáticamente. OK. |
| CI corre solo `go build` y no `go test` | media | Verificar workflow de CI (feature-020); agregar `go test ./internal/...` si falta. |

## Riesgos de integración / cross-repo

- Cross-repo: **NINGUNO**. Sin cambios FE. No hay contrato API nuevo que sincronizar.

## Riesgos de datos / migración

- **NINGUNO.** Cero migraciones. sqlite `:memory:` con `CREATE TABLE` inline por test; no toca DB real.

## Riesgos de archivos compartidos

- **NINGUNO.** Esta feature NO toca `wire/*`, `cmd/api/*`, `cmd/config/*`, `go.mod`, `go.sum`, `Makefile`,
  `internal/shared/**`. Cada test vive en su paquete; no hay hunks compartidos con otras features.

## Riesgo de extracción parcial

| riesgo | severidad | mitigación |
|---|---|---|
| Traer un `repository_tenant_test.go` sin que su repo de producción tenga tenancy | media | El paquete no compila. Detectar con `go test ./internal/<modulo>/`. Traer solo los tests cuyas features productoras ya estén. |
| Olvidar que `work-order/handler_test.go` es `M` (ya existe en develop) y dejar la versión vieja | media | `git checkout 777e5f6a -- internal/work-order/handler_test.go` sobreescribe con la versión correcta. Verificar que el stub use `HardDeleteWorkOrder` y `ListArchivedWorkOrders`. |

## Riesgo de mergear solo este repo / solo el otro

- **Solo BE (este repo):** correcto y suficiente — la feature es solo-BE. No deja al FE inconsistente.
- **Solo el otro repo (FE):** N/A, no hay parte FE.
- El único acoplamiento de orden es intra-BE (con 001/002/009), no cross-repo.
