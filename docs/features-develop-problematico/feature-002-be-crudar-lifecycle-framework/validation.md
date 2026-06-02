# validation.md — feature-002 · be-crudar-lifecycle-framework

## Checklist pre-PR

- [ ] go.mod tiene `github.com/prometheus/client_golang v1.23.2`,
      `github.com/devpablocristo/platform/observability/go v0.2.1`,
      `github.com/devpablocristo/platform/persistence/gorm/go v0.1.0`
      (+ indirects prometheus client_model/common/procfs).
- [ ] `gorm.io/driver/sqlite v1.6.0` presente (ya está en develop; lo usan tests).
- [ ] `go mod verify` OK; `go mod tidy` no deja diff inesperado.
- [ ] `git status` muestra SOLO los 17 paths de la feature (9 .go + 8 .sql)
      [+ go.mod/go.sum si se incluyen deps]. NINGÚN `internal/*/repository.go`
      ni `cmd/*`.
- [ ] `git diff --check` (sin trailing whitespace / conflict markers).
- [ ] Decisión tomada sobre el ORDEN de migraciones 227/228 vs 229/230.

## Tests sugeridos (BE)

```bash
# unit + e2e (sqlite in-memory, NO necesita Postgres)
go test ./internal/shared/lifecycle/...

# build de TODO el repo (el paquete no debe romper nada; nadie lo importa aún)
go build ./...
go vet ./internal/shared/lifecycle/...
```

Tests presentes que deben pasar:
- `cascade_test.go` — RunCascadeArchive/Restore, WouldOrphanActiveChildren.
- `archive_cleanup_test.go` — RunArchiveCleanup (dry-run y apply), reglas IA-*.
- `invariant_e2e_test.go` — invariante hierárquico customers→projects→fields→lots.
- `metrics_test.go` — RegisterMetrics + observeRejectedArchivedRef.

## Validación de migraciones (manual, en staging)

Idealmente sobre un dump de prod actualizado a develop (no sobre DB vacía):

1. Aplicar UP 227 → verificar tabla `archive_batches`, índices
   `idx_archive_batches_tenant_root`/`_created_at`, columnas `archive_batch_id`/
   `archive_origin_entity`/`archive_origin_id`/`archive_reason` + FK `*_archive_
   batch_id_fkey` (NOT VALID) en las 14 tablas core, índice
   `idx_<t>_archive_cause`, e índice parcial `ux_customers_tenant_actor_id`
   (con `WHERE ... deleted_at IS NULL`). Verificar
   `SELECT count(*) FROM actors WHERE deleted_at IS NOT NULL` ≈ los que tenían
   `archived_at`.
2. Aplicar UP 228 → mismas columnas en las 18 tablas restantes.
3. Aplicar UP 232 → `\d+ v4_report.workorder_list` muestra el COMMENT; idem
   funciones `v4_ssot.*`. Si falla → la vista/función no existe en destino.
4. Aplicar UP 233 → `\df public.assert_parent_active` existe; 14 triggers
   `trg_*_active_*` presentes (`SELECT tgname FROM pg_trigger WHERE NOT tgisinternal`).
5. Probar el invariante a nivel DB: intentar
   `UPDATE projects SET customer_id = <id-de-customer-archivado> WHERE id=...`
   debe fallar con SQLSTATE 23514 ("active projects row references archived
   customers").
6. Aplicar DOWN 233 → 232 → 228 → 227 en orden inverso; verificar limpieza.
   OJO: down de 227 NO recompone `actors.archived_at` desde `deleted_at`.

## Casos borde a verificar

- Restaurar un project NO debe revivir un field archivado por OTRA causa
  (ApplyCauseScope). Cubierto por tests Go; validable en staging con 009.
- Tablas que no existen en la DB destino: 227/228 las saltean (`to_regclass`);
  232/233 NO → fallan. Confirmar inventario de tablas antes de 233.
- `archive_batch_id` FK es `NOT VALID` → no valida filas existentes; OK para
  migración online.

## Qué revisar en UI / API / DB / env

- **UI**: nada (sin FE).
- **API**: nada propio; los endpoints archive/restore son feature 009.
- **DB**: ver migraciones arriba.
- **env**: la métrica usa namespace `ponti_backend` (hardcodeado en
  `cmd/api/main.go` de feature 009). No hay env nuevo en feature-002.

## Qué validar en el otro repo

- Nada. Confirmar en el cross-repo-map del FE que feature-002 figura como
  "sin cambios FE".

## Señales de incompletitud / incompatibilidad

- `go build ./...` falla por `prometheus`/`platform/observability`/`persistence`
  no resueltos → faltan deps (R1).
- `migrate up` rechaza 227/228 con "file does not exist" / "Dirty database" /
  out-of-order → R2 (orden).
- `CREATE TRIGGER` o `COMMENT ON` fallan → R4/R5 (schema destino incompleto).
- Si `go build ./...` falla por símbolos de actor/identity → se mezcló alcance
  de otra feature (R11).
