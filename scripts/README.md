# Scripts

Este directorio contiene solo herramientas operativas versionadas. Los archivos
generados no deben quedar commiteados.

## Raiz

| Script | Uso |
| --- | --- |
| `check_schema_guardrails.sh` | Validacion de schema usada por deploy dev/stg/prod. |
| `docker_cleanup.sh` | Limpieza interactiva y destructiva de Docker local. |
| `down_ponti_local.sh` | Baja el stack local Ponti completo. |
| `entrypoint.sh` | Entrypoint de la imagen Docker de produccion. |
| `run_ponti_local.sh` | Levanta backend, frontend y AI en local. |
| `smoke_release.sh` | Smoke test de release usado por deploys. |

## DB

| Script | Uso |
| --- | --- |
| `db_reset.sh` | Recrea la DB local del compose. |
| `db_migrate_up.sh` | Ejecuta migraciones v4 en local. |
| `db_validate.sh` + `db_validate.sql` | Validaciones de schema local. |
| `db_schema_snapshot.sh` | Genera `schema.snapshot.sql` local. No se versiona. |
| `db_schema_diff.sh` | Compara `schema.snapshot.sql` contra `schema.expected.sql`. |
| `db_adopt_baseline.sh` | Adopta baseline en ambientes existentes. |
| `reset-local-db-from-prod.sh` | Restaura datos de PROD en DB local. PROD es solo origen. |
| `actors_backfill_sync.sql` | Sync idempotente de actors luego de restore legacy. |
| `actors_golden_master.sql` | Golden master de migracion de actors. |
| `multi_tenant_golden_master.sql` | Golden master multi-tenant. |
| `tenant_isolation_audit.sql` | Auditoria read-only de aislamiento tenant. |
| `hardening_post_restore.sql` | Fuente SQL manual para hardening post import Cloud SQL. |

## Generados

- `scripts/db/schema.snapshot.sql`: generado por `make db-schema-snapshot`.
  Esta ignorado por git.
