# spec.md — feature-019 · be-local-tooling-db-scripts

- **id:** feature-019
- **nombre:** Local tooling & DB scripts
- **tipo:** infra (tooling / DevEx / scripts operativos)
- **repo:** Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **existe-en-FE/BE:** Solo BE. En FE (web) **no hay carpeta ni cambios** — en el cross-repo-map del FE figura como "sin cambios FE".
- **merge:** BE independiente.
- **rango fuente-de-verdad (diff):** `0972e565..777e5f6a`
- **SOURCE REF de extracción:** `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama destino:** `develop` (tip `003a9b8f`).

## Resumen

Conjunto de herramientas operativas versionadas bajo `scripts/` más el `Makefile`
de la raíz. Son los scripts "sobrevivientes" del set viejo (3-dot) más varios
nuevos que aparecieron con la migración a la arquitectura `platform/` (new-cns3):
DB (reset/migrate/diff/golden-master), auditorías de datos (data-audit,
tenant-isolation), lint de fugas de tenancy, smoke test de Companion, export de
conversaciones de IA y orquestación local del stack (`run/down_ponti_local.sh`).

## Objetivo

Que un dev/operador pueda, en local:
1. Resetear la DB local y cargar datos data-only desde PROD (read-only) de forma segura.
2. Re-correr el backfill/sync idempotente de actors tras un restore legacy.
3. Validar invariantes de datos (jerarquía archivada, aislamiento de tenant) en modo read-only.
4. Detectar regresiones del refactor de multi-tenancy en CI (`lint-tenant-leaks.sh`).
5. Levantar/bajar el stack local (`core` + `web`; `axis` se gestiona aparte).
6. Hacer smoke test del cliente Companion y exportar conversaciones de IA.

## Problema

El set de scripts viejo asumía la topología `ponti-backend` + `ponti-frontend` +
`ponti-ai` y la descarga desde **STAGING**. Con new-cns3 la topología pasó a
`core` + `web` + `axis` (repo paralelo), el origen de datos pasó a **PROD
read-only**, y se introdujeron tablas/conceptos nuevos (actors, tenancy estricta)
que necesitan golden masters y auditorías propias. Además había artefactos
generados commiteados (`schema.snapshot.sql`, 6589 líneas) y scripts peligrosos
contra GCP (`db_force_reset_gcp`, `db_gcp_reset_and_load_local`) que se eliminaron.

## Alcance en este repo (BE)

Propios de la feature (tooling puro, sin tocar código de la app):
- `scripts/README.md`, `scripts/data-audit/*`, `scripts/db/*` (nuevos + modificados),
  `scripts/lint-tenant-leaks.sh`, `scripts/export-ai-conversations.sh`,
  `scripts/smoke-companion/main.go`, `scripts/run_ponti_local.sh`, `scripts/down_ponti_local.sh`.
- `Makefile` — **COMPARTIDO**: su diff mezcla hunks de esta feature con hunks de
  otras (OpenAPI/024, rename core→platform/001, lint v2/020-022, baja de seed). Ver file-list.md.

## Alcance en el otro repo (FE / web)

Ninguno. Feature solo-BE. Documentar en el FE como "sin cambios FE".
(El `Makefile` y `run_ponti_local.sh` *mencionan* `web/` y `axis/` por texto/paths,
pero no se modifica ningún archivo del repo web.)

## Fuera de alcance

- Código de la aplicación Go (`internal/**`, `cmd/**`) — NO se toca. Los scripts
  solo lo *invocan* (`cmd/api`, `cmd/archive-cleanup`).
- Migraciones (`migrations_v4/**`) — NO forman parte de esta feature (los `.sql`
  de `scripts/db/` son auditorías/golden-masters/backfills operativos, no migraciones del schema).
- Hunks del `Makefile` que pertenecen a otras features (OpenAPI codegen → 024;
  rename `core/*`→`platform/*` → 001; lint `golangci-lint v2.11.4` → 020/022).
- CI workflows (020), build/deploy config (021), git hooks (022).

## Comportamiento esperado

- `make reset-local-db-from-prod` → levanta `ponti-db`, corre
  `scripts/db/reset-local-db-from-prod.sh` (data-only desde PROD, guardas de seguridad
  que bloquean destino no-localhost, `DRY_RUN=1` soportado, `MIGRATE_TARGET_VERSION=224`
  por defecto para dumps legacy).
- `make actors-backfill-sync` → corre `scripts/db/actors_backfill_sync.sql` (idempotente).
- `./scripts/lint-tenant-leaks.sh` → exit 0 si limpio, 1 si hay fugas de tenancy.
- `make up-ponti-local` / `down-ponti-local` → orquestan `core` + `web` (axis aparte).
- `go run ./scripts/smoke-companion` → smoke del cliente Companion contra axis local.
- `psql ... -f scripts/db/tenant_isolation_audit.sql` → 0 filas antes de activar strict mode.

## Estado en dp~1 (`777e5f6a`)

Completo y coherente como set de tooling. Todos los archivos nuevos existen y los
modificados ya apuntan a la topología nueva. Bajo riesgo. Los `.sql` golden-master
y de auditoría asumen que el schema/tablas de las features de las que dependen
(actors/007, tenancy/003) ya están aplicadas en la DB sobre la que corren — pero
**eso es una dependencia de runtime de la DB, no de compilación del repo**.

## Criterios de aceptación

1. `make help`/`make -n` resuelve los targets nuevos sin referenciar targets borrados.
2. `bash -n` (syntax check) pasa en todos los `.sh` extraídos.
3. `./scripts/lint-tenant-leaks.sh` corre sin error de sintaxis (resultado funcional
   depende de feature-001/003 ya mergeadas).
4. `go build ./scripts/smoke-companion` compila (depende de `internal/axis` → feature-012).
5. No queda commiteado ningún artefacto generado (`schema.snapshot.sql` sigue gitignored).
6. No reaparecen los scripts GCP peligrosos eliminados.

## Endpoints / Modelos / UI / DB / Tests afectados

- **Endpoints/rutas:** ninguno (no toca routers ni handlers). `smoke-companion`
  *consume* `POST /v1/chat` del servicio axis (externo), no lo define.
- **Modelos/DTOs/tipos:** ninguno propio. `smoke-companion/main.go` usa tipos de
  `internal/axis` (`axis.Config`, `axis.CallContext`, `axis.ChatRequest`, `axis.NewCompanionClient`).
- **UI/componentes:** ninguno.
- **DB:** SQL operativo read-only/backfill (no DDL de migración):
  `archived_invariants.sql`, `tenant_isolation_audit.sql`, `actors_golden_master.sql`,
  `multi_tenant_golden_master.sql`, `actors_backfill_sync.sql`.
- **Tests:** ninguno de Go (`*_test.go`). Validación = `bash -n`, `make -n`, ejecución manual.

## Dependencias

**Intra-repo (DEPENDE DE: ninguna en sentido fuerte; soft en runtime):**
- feature-001 (be-platform-tenancy-refactor) — `lint-tenant-leaks.sh` referencia
  `platform/persistence/gorm/go/tenancy`, `domainerr.TenantMissing()`, eliminación de `authz.MaybeTenantScope`.
- feature-003 (be-multitenant-db-hardening) — `tenant_isolation_audit.sql`,
  `multi_tenant_golden_master.sql` asumen columnas `tenant_id`.
- feature-007 (actor-system) — `actors_*.sql` asumen tablas `actors`, `actor_roles`,
  `legacy_actor_map`, `normalize_actor_name()`.
- feature-012 (ai-companion-integration) — `smoke-companion/main.go` importa `internal/axis`.
- feature-018 (data-integrity-admin) — `data-audit/README.md` documenta `cmd/archive-cleanup`.

**Cross-repo:** Ninguna (solo-BE).

## Riesgos

- **Funcional:** los `.sql` y `lint-tenant-leaks.sh` *parecen* correr pero dan
  resultados vacíos/falsos si la DB o el código de las features dependientes no
  están presentes. No rompen el build; rompen la *utilidad*.
- **Técnico:** el `Makefile` es compartido; extraerlo entero arrastra hunks de
  otras features (OpenAPI, rename platform, lint v2). Hay que extraer por hunks.
- **Datos:** `reset-local-db-from-prod.sh` toca PROD como origen (read-only) y la DB
  local como destino destructivo. Tiene guardas (`DB_HOST` debe ser localhost), pero
  es el script más peligroso del set si se mal-configura.

## DECISIÓN recomendada

**Extraer tal cual los archivos propios** (scripts y READMEs), y **partir el `Makefile`
en partial-hunks** (traer solo los hunks de tooling-DB; dejar OpenAPI/rename/lint a sus features).
No bloquear el merge por las dependencias soft: son tooling y no rompen compilación.
Marcar `smoke-companion/main.go` como **manual-port/condicionado**: solo compila si
feature-012 (`internal/axis`) ya está en `develop`. Si no, traerlo igual (no rompe `go build ./...`
del binario principal, pero `go build ./scripts/...` fallaría — ver risks.md).
