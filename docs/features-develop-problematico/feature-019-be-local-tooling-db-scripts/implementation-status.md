# implementation-status.md — feature-019 · be-local-tooling-db-scripts

## Estado global

**Completa** como set de tooling en SOURCE (`777e5f6a`). Todos los archivos nuevos
existen y los modificados ya apuntan a la topología new-cns3 (`core`+`web`+`axis`)
y a PROD-read-only como origen.

- **% completitud:** ~95%. Lo que resta es de *extracción* (partir el Makefile, condicionar smoke-companion), no de implementación.

## Estado en este repo (BE)

| componente | estado | nota |
| --- | --- | --- |
| `scripts/README.md`, `data-audit/README.md` | completo | docs índice |
| `archived_invariants.sql` | completo | IA-1..IA-14 read-only |
| `actors_backfill_sync.sql` | completo | idempotente, TRUNCATE+reinsert por tenant |
| `actors_golden_master.sql`, `multi_tenant_golden_master.sql` | completo | read-only, diff=0 esperado |
| `tenant_isolation_audit.sql` | completo | read-only, 0 filas antes de strict mode |
| `lint-tenant-leaks.sh` | completo | grep CI guard |
| `export-ai-conversations.sh` | completo | psql read-only a axis DSN |
| `run_ponti_local.sh` / `down_ponti_local.sh` | completo | reescritos a core/web; axis aparte |
| `db_migrate_up.sh` | completo | comentario v2→v4 |
| `db_schema_diff.sh` | completo | reorden de guardas |
| `reset-local-db-from-prod.sh` | completo | guardas localhost, DRY_RUN, MIGRATE_TARGET_VERSION=224 |
| `smoke-companion/main.go` | completo PERO condicionado | compila solo con `internal/axis` (012) presente |
| `Makefile` (hunks de 019) | completo | requiere extracción parcial |

## Estado en el otro repo (FE / web)

No aplica. Sin cambios FE.

## Tests

- No hay tests de Go (`*_test.go`) en la feature.
- Validación = `bash -n`, `make -n`, `go build ./scripts/smoke-companion`, ejecución manual.

## Pendientes / bugs

### BLOQUEANTE para mergear
- **Ninguno** si se excluye/condiciona `smoke-companion/main.go`. Si se incluye sin
  feature-012 en `develop`, `go build ./scripts/...` falla → bloqueante para CI que compile todo.

### Mejora futura
- Enganchar `lint-tenant-leaks.sh` a `make lint` / pipeline CI (coordinar feature-020).
- Follow-up para `smoke-companion/main.go` si se posterga por falta de 012.

### Deuda aceptable
- Los `.sql` golden-master/auditoría dependen del schema en runtime; documentado, no hay validación automática de "se corrieron".
- `reset-local-db-from-prod.sh` toca PROD como origen: guardas presentes pero requiere disciplina del operador (ver MEMORY: el set viejo asumía numeración de migraciones nueva; este script ya fija `MIGRATE_TARGET_VERSION=224` para dumps legacy).

### Duda humana
- Confirmar qué feature "posee" el hunk `cmd/`→`cmd/api` del Makefile (¿021? ¿023?).
- Confirmar si la baja de `seed`/`seed-dashboard`/`select-ponti-*` del Makefile debe ir en esta
  feature o en config-modularization (005). Recomendación: NO traerla en 019.
