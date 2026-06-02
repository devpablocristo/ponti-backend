# feature-018 · data-integrity-admin · implementation-status (BE)

## Estado global

- **Estado**: COMPLETA en SOURCE (`777e5f6a`), pero **NO portable en aislamiento** sobre develop.
- **% completitud (módulo en sí)**: ~100% en SOURCE.
- **% completitud (extracción sobre develop sin dependencias)**: ~60% — falta lo compartido (authz + 4 RAW + wire), que no está en mi flist.

## Estado en ESTE repo (BE)

- En `develop-problematico~1` (777e5f6a): módulo `internal/data-integrity` reescrito, coherente, con tests; compila contra ese árbol (que ya tiene authz + RAW methods + wire nuevo).
- En `develop` (tip 003a9b8f): el módulo está en su versión VIEJA (17 controles, 1068 líneas, `StockRepositoryPort`, timeout 8m, import `core/...domainerr`). Las dependencias nuevas (authz, 4 RAW, wire nuevo) **no existen**.

### Detalle de mi flist

| archivo | estado en SOURCE | observación |
|---|---|---|
| `usecases.go` | completo | 5 controles + `buildCheck`; paralelo con WaitGroup; tolerancia 1 USD |
| `handler.go` | completo | ruta intacta; timeout 30s; sin `GetProtected()` |
| `handler/dto/integrity_check.go` | completo | DTO con 3 campos nuevos |
| `usecases/domain/types.go` | completo | dominio sin json tags |
| `handler_test.go` | completo (nuevo) | 2 tests |
| `usecases_test.go` | completo | 6 tests |
| `usecases_tenant_test.go` | completo (nuevo) | tenant propagation a 6 repos |
| `usecases_mock_test.go` | completo | mocks de 6 ports |

## Estado en el OTRO repo (FE)

- DESCONOCIDO desde este paquete (no inspeccionado). Existe `feature-018` FE con `pages/admin/data-integrity` + `useDatabase`. Debe alinearse al DTO nuevo. Ver paquete FE.

## Tests

- Cobertura del módulo: handler (2), usecases (6), tenant (1) → buena cobertura de la lógica nueva.
- Los tests son whole-file y autocontenidos (usan mocks gomock + stubs); **no** dependen de los repos compartidos, así que pasan apenas el paquete compile.
- **OJO**: que los tests del paquete pasen NO garantiza que `go build ./...` del repo entero compile — el build necesita authz + 4 RAW + wire.

## Pendientes

### BLOQUEANTE para mergear (este PR no debe mergear sin esto)

- `internal/shared/authz` presente en develop (de 001/003).
- 4 métodos RAW (report/supply/project/lot) traídos por partial-hunks.
- `wire/data_integrity_providers.go` re-cableado + `wire/wire_gen.go` regenerado.
- `go build ./...` verde.

### Mejora futura

- Tenantizar `GetRawDirectCost` (work-order) para igualar el patrón de los otros 4 RAW.
- Activar/emitir `WARNING/SKIPPED` (hoy el dominio los documenta pero `buildCheck` solo emite `OK/ERROR`).
- `Recommendation` está en el DTO/dominio pero `buildCheck` no lo setea (queda vacío) — completar mensajes de recomendación.

### Deuda aceptable

- `CheckType` siempre es `STRONG` (constante `checkTypeStrong`); las constantes `WEAK/FORMULA_ALIGNMENT` mencionadas en comentarios no se usan aún.
- Comentarios del DTO/dominio dicen "17 controles" pero la implementación tiene 5 (residuo del refactor). Cosmético.

### Duda humana

- ¿La limpieza de json-tags de `usecases/domain/types.go` va con 018 o con 027? Recomendación: con 018 (es local al módulo).
- ¿La tenantización de `GetRawDirectCost` la dueña es 018 o tenancy?

## Bugs / inconsistencias observadas

- Ninguno funcional dentro del módulo. La única "trampa" es de extracción: el módulo compila en SOURCE pero referencia símbolos que no están en develop (authz + 4 RAW). Es una incompletitud de scope, no un bug de código.
