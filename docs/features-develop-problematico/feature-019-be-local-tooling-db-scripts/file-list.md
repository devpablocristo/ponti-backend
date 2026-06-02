# file-list.md — feature-019 · be-local-tooling-db-scripts

Lista autoritativa: `/tmp/flists/be-019.txt` (19 paths). STATUS: A=created, M=modified, D=deleted.
Diff de verdad: `git -C <core> diff 0972e565..777e5f6a -- <path>`.

Leyenda extracción: `whole-file` = traer el archivo completo del SOURCE · `partial-hunks`
= traer solo algunos hunks (archivo compartido) · `manual-port` = revisar/condicionar a otra feature
· `do-not-extract-yet` = no traer todavía.

## Propios (núcleo de la feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `scripts/README.md` | A | doc | índice de scripts versionados (raíz + db + generados) | whole-file | doc puro, sin deps | bajo | alta |
| `scripts/data-audit/README.md` | A | doc | guía de auditoría de invariantes archivadas; documenta `cmd/archive-cleanup` (IA-1..IA-14) | whole-file | doc; referencia soft a 018 | bajo | alta |
| `scripts/data-audit/archived_invariants.sql` | A | SQL (read-only) | detecta hijos activos bajo padre archivado (IA-1..IA-14) | whole-file | SQL read-only, sin DDL | bajo | alta |
| `scripts/db/actors_backfill_sync.sql` | A | SQL (backfill idempotente) | re-sync 1:1 de actors tras restore legacy; TRUNCATE + re-insert por tenant | whole-file | idempotente; asume tablas actors/007 en DB | medio | alta |
| `scripts/db/actors_golden_master.sql` | A | SQL (read-only) | golden master de migración de actors (diff debe ser 0) | whole-file | read-only; asume `legacy_actor_map`/007 | bajo | alta |
| `scripts/db/multi_tenant_golden_master.sql` | A | SQL (read-only) | golden master multi-tenant (group by tenant/project/campaign/field) | whole-file | read-only; asume `tenant_id`/003 | bajo | alta |
| `scripts/db/tenant_isolation_audit.sql` | A | SQL (read-only) | auditoría de aislamiento tenant; 0 filas antes de strict mode | whole-file | read-only; asume tenancy/003 | bajo | alta |
| `scripts/lint-tenant-leaks.sh` | A | shell (CI guard) | falla si hay fugas de tenancy (string literal, authz.* eliminados, WHERE inline) | whole-file | grep-only; asume `platform/.../tenancy`/001 | bajo | alta |
| `scripts/export-ai-conversations.sh` | A | shell (read-only) | exporta `ai_conversations` de axis a JSONL+summary; read-only | whole-file | psql read-only; DSN externo | bajo | alta |
| `scripts/smoke-companion/main.go` | A | Go (script) | smoke del cliente Companion; importa `internal/axis` | manual-port | **compila solo si feature-012 está en develop** | medio | alta |
| `scripts/run_ponti_local.sh` | M | shell | levanta stack local `core`+`web` (axis aparte); reemplaza topología vieja | whole-file | reescritura completa de paths; sin deps de código | bajo | alta |
| `scripts/down_ponti_local.sh` | M | shell | baja stack local `core`+`web`; axis aparte | whole-file | reescritura; sin deps | bajo | alta |
| `scripts/db/db_migrate_up.sh` | M | shell | corre migraciones v4 (antes decía v2) | whole-file | cambio mínimo de comentario v2→v4 | bajo | alta |
| `scripts/db/db_schema_diff.sh` | M | shell | reordena chequeos: exige snapshot antes que expected | whole-file | reorden de guardas; trivial | bajo | alta |
| `scripts/db/reset-local-db-from-prod.sh` | M | shell (destructivo local) | reset DB local + migrate + data-only desde PROD read-only | whole-file | guardas localhost, DRY_RUN, MIGRATE_TARGET_VERSION=224 | medio | alta |

## Compartidos (partial-hunks)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `Makefile` | M | build/tooling | targets de tooling DB/stack | **partial-hunks** | diff mezcla varias features (ver abajo) | alto | media |

**Hunks del `Makefile` que SÍ pertenecen a feature-019 (traer):**
- Target `reset-local-db-from-prod` (reemplaza `db-staging-to-local`/`db-reset-from-staging`).
- Target `actors-backfill-sync` (reemplaza `staging-db-2-dev-db`).
- `up-ponti-local` / `down-ponti-local` → wiring a `run_ponti_local.sh` / `down_ponti_local.sh` (texto core/web/axis).
- Eliminación de targets GCP peligrosos: `db-force-reset-gcp`, `db-gcp-reset-and-load-local`.
- Eliminación de `db-staging-to-local`, `db-reset-from-staging`, `staging-db-2-dev-db`.
- Header `.PHONY` ajustado a los targets anteriores.

**Hunks del `Makefile` que NO son de feature-019 (dejar a sus features):**
- Target `openapi:` (swag) + comentario → **feature-024 (openapi-and-docs)**.
- `lint:` cambio a `golangci-lint v2.11.4` → **feature-020/022**.
- `bin-build`/`run` `cmd/` → `cmd/api` → tocaría build (021/023); confirmar con dueño de cmd.
- Eliminación de `select-ponti-stg-local`/`dev-local`, `seed`, `seed-dashboard` → cleanup de entornos/seed; coordinar con 005/config.
- Reemplazos de texto `core/*` → `platform/*` en warnings → **feature-001 (rename)**.

## Requeridos por dependencia (NO se extraen aquí; solo deben existir en develop)

| recurso | provisto por | por qué lo necesita 019 |
| --- | --- | --- |
| `internal/axis` (paquete) | feature-012 | `smoke-companion/main.go` lo importa |
| `cmd/archive-cleanup` | feature-018 | `data-audit/README.md` lo documenta (no lo compila) |
| tablas `actors`/`actor_roles`/`legacy_actor_map`, `normalize_actor_name()` | feature-007 | `actors_*.sql` (runtime DB, no build) |
| columnas `tenant_id`, strict mode | feature-003 | `tenant_isolation_audit.sql`, `multi_tenant_golden_master.sql` (runtime DB) |
| `platform/.../tenancy`, `domainerr.TenantMissing()` | feature-001 | `lint-tenant-leaks.sh` (grep, no build) |

## Dudosos

| path | nota |
| --- | --- |
| `scripts/db/db_migrate_up.sh` | cambio mínimo (v2→v4 en comentario); confirmar que no choca con migraciones de 003. Confianza alta de que es trivial. |
| `Makefile` hunk `cmd/` → `cmd/api` | puede pertenecer a un cleanup de `cmd/` (021/023). Si al partir hunks aparece conflicto, dejarlo a esa feature. |

## NO traer todavía (deleted en SOURCE — son borrados, no contenido a portar)

| path | status | qué hacer | motivo |
| --- | --- | --- | --- |
| `scripts/db/repair_stocks_investor_granularity.sql` | D | replicar el `git rm` (borrarlo en develop si existe) | one-shot ya aplicado de un fix de stocks histórico; no es tooling vivo |
| `scripts/db/schema.snapshot.sql` | D | replicar el `git rm` + confirmar `.gitignore` lo cubre | artefacto generado (6589 líneas) que no debe estar versionado; lo genera `make db-schema-snapshot` |
