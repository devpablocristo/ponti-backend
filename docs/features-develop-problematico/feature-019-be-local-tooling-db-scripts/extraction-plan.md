# extraction-plan.md — feature-019 · be-local-tooling-db-scripts

- **repo:** `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip `003a9b8f`)
- **SOURCE:** `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip = restore vacío).
- **rama sugerida:** `pr/feature-019-be-local-tooling-db-scripts-be`
- **merge:** BE independiente. Sin coordinación cross-repo (solo-BE).

> Todos los comandos `git` de abajo son **SUGERENCIAS para un humano**. Este doc
> no ejecuta nada que mute el repo.

## PR title

`infra(be): local tooling & DB scripts (reset-from-prod, golden masters, tenant-leak lint, smoke-companion)`

## PR description (borrador)

> Trae el set de tooling operativo versionado (scripts/) y los hunks de tooling-DB del Makefile,
> alineado a la topología new-cns3 (`core` + `web` + `axis`) y a PROD-read-only como origen de datos.
> Incluye: reset-local-db-from-prod con guardas de seguridad, backfill/sync idempotente de actors,
> golden masters (actors / multi-tenant), auditorías read-only (archived invariants, tenant isolation),
> lint anti-fugas de tenancy para CI, smoke test del cliente Companion y export de conversaciones de IA.
> Elimina scripts GCP destructivos (`db_force_reset_gcp`, `db_gcp_reset_and_load_local`),
> el dump data-only desde STAGING y el snapshot de schema generado que estaba commiteado.
> Solo-BE: sin cambios en el repo web. No toca código de la app ni migraciones.

## Orden de extracción

### 0. Pre-requisitos en `develop` (verificar, no extraer aquí)
Estas features deberían estar ya en `develop` para que el tooling sea *útil* y `smoke-companion` *compile*:
- feature-012 (`internal/axis`) → necesario para `go build ./scripts/smoke-companion`.
- feature-001/003/007 → necesarios solo en runtime de la DB (no para compilar/mergear este PR).
Si feature-012 NO está aún: ver "qué podría romperse" abajo (opción: postergar solo `smoke-companion/main.go`).

### 1. Crear rama
```
git checkout develop
git pull
git checkout -b pr/feature-019-be-local-tooling-db-scripts-be
```

### 2. Traer archivos enteros (created/modified, propios)
```
git checkout develop-problematico~1 -- \
  scripts/README.md \
  scripts/data-audit/README.md \
  scripts/data-audit/archived_invariants.sql \
  scripts/db/actors_backfill_sync.sql \
  scripts/db/actors_golden_master.sql \
  scripts/db/multi_tenant_golden_master.sql \
  scripts/db/tenant_isolation_audit.sql \
  scripts/lint-tenant-leaks.sh \
  scripts/export-ai-conversations.sh \
  scripts/run_ponti_local.sh \
  scripts/down_ponti_local.sh \
  scripts/db/db_migrate_up.sh \
  scripts/db/db_schema_diff.sh \
  scripts/db/reset-local-db-from-prod.sh
```

### 3. smoke-companion (condicionado a feature-012)
- Si `internal/axis` ya existe en `develop`:
```
git checkout develop-problematico~1 -- scripts/smoke-companion/main.go
go build ./scripts/smoke-companion   # debe compilar
```
- Si NO existe aún: **no traer** `scripts/smoke-companion/main.go` en este PR;
  abrir follow-up tras mergear 012. Anotarlo en el PR.

### 4. Replicar borrados (deleted en SOURCE)
```
git rm scripts/db/repair_stocks_investor_granularity.sql   # si existe en develop
git rm scripts/db/schema.snapshot.sql                       # si existe en develop
```
Confirmar que `.gitignore` ya ignora `scripts/db/schema.snapshot.sql` (lo genera `make db-schema-snapshot`).

### 5. Makefile — PARTIAL HUNKS (no traer entero)
```
git restore -p --source=develop-problematico~1 -- Makefile
```
Aceptar SOLO los hunks de tooling-DB/stack (ver file-list.md → "Hunks que SÍ pertenecen"):
- `reset-local-db-from-prod`, `actors-backfill-sync`, `up/down-ponti-local`,
  eliminación de targets GCP/staging, ajuste de `.PHONY`.
RECHAZAR los hunks de otras features:
- `openapi:` (→024), `lint: golangci-lint v2.11.4` (→020/022),
  rename `core/*`→`platform/*` (→001), baja de `select-ponti-*`/`seed`/`seed-dashboard` (coordinar 005),
  `bin-build`/`run` `cmd/`→`cmd/api` (coordinar 021/023).

Si partir hunks resulta inviable por solapamiento, alternativa: editar el `Makefile`
de `develop` a mano aplicando solo los targets de la lista, y dejar el resto.

### 6. Permisos de ejecución
`reset-local-db-from-prod.sh` cambió a modo `100755` en SOURCE. Verificar bit +x:
```
chmod +x scripts/db/reset-local-db-from-prod.sh scripts/lint-tenant-leaks.sh \
         scripts/export-ai-conversations.sh scripts/run_ponti_local.sh scripts/down_ponti_local.sh
git update-index --chmod=+x scripts/db/reset-local-db-from-prod.sh
```

## Archivos enteros vs parciales

- **Enteros:** todos los `scripts/**` (14 propios + smoke condicionado) y los 2 borrados.
- **Parciales:** solo `Makefile`.

## Migraciones / tests a incluir

- **Migraciones:** ninguna (los `.sql` son operativos/auditoría, NO migraciones del schema).
- **Tests:** ninguno de Go. Validación por `bash -n`, `make -n` y `go build ./scripts/...`.

## Dependencias previas

Ninguna **bloqueante de merge**. Soft (runtime/compilación de smoke):
- feature-012 antes de traer `smoke-companion/main.go`.
- 001/003/007 antes de *usar* (no de mergear) los `.sql`/lint.

## Coordinación con el otro repo

No aplica (solo-BE). Mencionar en cross-repo-map del FE: "sin cambios FE".

## Comandos git SUGERIDOS (resumen)

```
git checkout develop
git checkout -b pr/feature-019-be-local-tooling-db-scripts-be
git checkout develop-problematico~1 -- <paths enteros del paso 2>
git restore -p --source=develop-problematico~1 -- Makefile      # solo hunks de 019
git rm scripts/db/repair_stocks_investor_granularity.sql
git rm scripts/db/schema.snapshot.sql
git diff --check                                                  # sin trailing whitespace / conflict markers
git diff --cached --stat
```

## Qué NO traer

- Hunks del `Makefile` de openapi/lint-v2/rename-platform/seed/cmd-api (cada uno a su feature).
- `develop-problematico` (tip vacío) como fuente — usar SIEMPRE `~1`.
- Recrear `repair_stocks_investor_granularity.sql` ni `schema.snapshot.sql`.

## Qué podría romperse

- `go build ./scripts/smoke-companion` falla si `internal/axis` (012) no está → por eso es condicionado.
- Si se trae el `Makefile` entero por error: se arrastra `openapi:`, rename platform y lint-v2,
  rompiendo el aislamiento de features y posiblemente `make lint` si golangci v2 no está disponible.
- `reset-local-db-from-prod.sh` mal-configurado (DB_HOST != localhost) — el propio script lo bloquea,
  pero validar la guarda tras extraer.

## Cómo detectar extracción incompleta

- `git diff develop-problematico~1 -- scripts/ Makefile` debería mostrar SOLO los hunks
  no-019 del Makefile pendientes (todo lo demás idéntico).
- `make -n reset-local-db-from-prod` / `make -n actors-backfill-sync` deben resolver sin error.
- `grep -rn "db-staging-to-local\|db-force-reset-gcp\|staging-db-2-dev-db" Makefile` → 0 hits.
- `ls scripts/db/reset-local-db-from-prod.sh scripts/smoke-companion/main.go` presentes.

## Qué validar antes del PR

- `bash -n` en cada `.sh` extraído.
- `make -n` para los targets nuevos.
- `go build ./scripts/smoke-companion` (si se incluyó).
- `git diff --check` limpio.
- Bit +x en los scripts ejecutables.

## Qué hacer después de mergear

- Si se postergó `smoke-companion/main.go`: abrir follow-up cuando 012 esté en develop.
- Agregar `./scripts/lint-tenant-leaks.sh` al pipeline/`make lint` (coordinar con feature-020 CI).
- Comunicar al equipo el cambio de origen STAGING→PROD-read-only y la nueva topología core/web/axis.
