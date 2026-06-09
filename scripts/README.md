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
| `run_ponti_local.sh` | Levanta core, web y axis en local. |
| `smoke_release.sh` | Smoke test de release usado por deploys. |

## Axis / IA

Estos scripts validan Ponti como producto real consumidor de Axis. No levantan
servicios: Axis, Nexus y Ponti API deben estar corriendo antes de ejecutarlos.

Env vars base:

| Variable | Uso |
| --- | --- |
| `AXIS_COMPANION_BASE_URL` | URL de Companion, por ejemplo `http://localhost:18085`. |
| `AXIS_COMPANION_API_KEY` | API key operativa de Companion. |
| `NEXUS_BASE_URL` | URL de Nexus para el smoke aprobado. |
| `NEXUS_ADMIN_API_KEY` | API key de Nexus con scopes admin/result/approval. |
| `PONTI_BASE_URL` | URL de Ponti API, por ejemplo `http://localhost:8080`. |
| `PONTI_ORG_ID` | `auth_tenants.id`; se usa como `org_id` Axis. |
| `PONTI_AXIS_API_KEY` | Bearer server-to-server que Axis usa contra Ponti. |
| `PONTI_API_KEY` | API key local de Ponti para chat/previews cuando el smoke no usa bearer Axis. |
| `PONTI_PROJECT_ID` | Proyecto de Ponti owned por el tenant del smoke. |
| `PONTI_FIELD_ID` | Campo opcional para preview de OT. |
| `PONTI_CAMPAIGN_ID` | Campania opcional para preview de OT. |
| `PONTI_SUPPLY_ID` | Insumo opcional para preview de ajuste de stock. |

Orden recomendado para diagnostico incremental:

1. `scripts/axis/onboard-ponti.sh`
2. `scripts/axis/smoke-ponti-axis-readonly.sh`
3. `scripts/axis/smoke-ponti-axis-draft-actions.sh`
4. `scripts/axis/smoke-ponti-axis-draft-previews.sh`
5. `scripts/axis/smoke-ponti-axis-nexus-approved-draft.sh`
6. `scripts/axis/smoke-ponti-axis-chat.sh`

Ejecucion completa:

```bash
make smoke-axis-all
```

Targets disponibles:

| Target | Uso |
| --- | --- |
| `make smoke-axis` | Valida onboarding, discovery y ejecucion read-only por Axis. |
| `make smoke-axis-chat` | Valida chat Ponti -> Axis manteniendo contrato web legacy. |
| `make smoke-axis-governance` | Valida draft actions, previews y ejecucion aprobada por Nexus. |
| `make smoke-axis-all` | Ejecuta todos los smokes Axis/Ponti en orden y falla en el primer error. |

Ejemplo local completo:

```bash
export AXIS_COMPANION_BASE_URL=http://localhost:18085
export AXIS_COMPANION_API_KEY=local-dev-companion-api-key
export NEXUS_BASE_URL=http://localhost:18086
export NEXUS_ADMIN_API_KEY=local-dev-nexus-api-key
export PONTI_BASE_URL=http://localhost:8080
export PONTI_ORG_ID=<auth_tenants.id>
export PONTI_AXIS_API_KEY=local-dev-ponti-axis-api-key
export PONTI_API_KEY=<ponti-api-key>
export PONTI_PROJECT_ID=<project-id>
export PONTI_FIELD_ID=<field-id>
export PONTI_CAMPAIGN_ID=<campaign-id>
export PONTI_SUPPLY_ID=<supply-id>

make smoke-axis-all
```

El smoke Nexus aprobado es idempotente para uso local: crea o actualiza el
`action_type` `agent.capability.invoke` y la policy local que requiere
aprobacion para `target_system=ponti`. No ejecuta writes reales en Ponti; la
capability aprobada devuelve preview con `write_performed=false`.

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
