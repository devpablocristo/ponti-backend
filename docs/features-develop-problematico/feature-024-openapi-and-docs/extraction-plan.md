# extraction-plan.md — feature-024 · openapi-and-docs (BE)

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip restore/vacío).
- **rango de validación**: `0972e565..777e5f6a`
- **rama sugerida**: `pr/feature-024-openapi-and-docs-be`

## PR title

`docs(be): OpenAPI pipeline + arquitectura/CRUDAR/errores/observabilidad/tenancy`

## PR description (sugerida)

> Documentación-only del backend. Sin cambios de código, deps, migraciones ni config.
>
> **Nuevos (17)**: `CLAUDE.md`, `CRUDAR_PLAN.md`, `docs/OPENAPI.md`, `docs/ERROR_CATALOG.md`, `docs/OBSERVABILITY.md`, `docs/crudar-lifecycle.md`, `docs/archive-restore-policy.md`, `docs/entity-capabilities.md`, `docs/customers-projects-lifecycle.md`, `docs/DATA_INTEGRITY_CONTRACT.md`, `docs/MULTI_TENANT_100_EVIDENCE.md`, `docs/BACKEND_CLEANUP_AUDIT.md`, `docs/audit-custom-errors.md`, `docs/projects-archive-audit.md`, `docs/openapi/{openapi.yaml,swagger.yaml,swagger.json}`.
> **Modificados (3, partial)**: `docs/ARCHITECTURE.md` (secciones hex/observabilidad/lifecycle/errores/seguridad), `README.md` y `docs/README.md` (sólo hunks de doc, ver nota de coordinación con 019/021).
>
> El spec OpenAPI es un **piloto de 2 endpoints** (`/me/context`, `/data-integrity/...`); los ~48 handlers restantes quedan sin anotar (declarado en `docs/OPENAPI.md`).
>
> Cross-repo: el FE feature-024 consume `docs/openapi/swagger.yaml` vía `yarn codegen:openapi`. No bloqueante para mergear este BE.

## Pasos ordenados

1. **Partir de develop limpio**
   - `git -C <repo> checkout develop && git pull` (asegurar tip `003a9b8f` o posterior).
   - `git -C <repo> checkout -b pr/feature-024-openapi-and-docs-be`

2. **Traer los 17 archivos nuevos enteros** (sin riesgo, no existen en develop):
   - `git -C <repo> checkout 777e5f6a -- CLAUDE.md CRUDAR_PLAN.md`
   - `git -C <repo> checkout 777e5f6a -- docs/OPENAPI.md docs/ERROR_CATALOG.md docs/OBSERVABILITY.md docs/crudar-lifecycle.md docs/archive-restore-policy.md docs/entity-capabilities.md docs/customers-projects-lifecycle.md docs/DATA_INTEGRITY_CONTRACT.md docs/MULTI_TENANT_100_EVIDENCE.md docs/BACKEND_CLEANUP_AUDIT.md docs/audit-custom-errors.md docs/projects-archive-audit.md`
   - `git -C <repo> checkout 777e5f6a -- docs/openapi/openapi.yaml docs/openapi/swagger.yaml docs/openapi/swagger.json`

3. **`docs/ARCHITECTURE.md` (modificado) — partial/whole según conflicto**:
   - El diff de 024 es mayoritariamente contenido nuevo apendizado y un reescrito del TLDR. Si develop NO tocó ARCHITECTURE.md desde `0972e565`, se puede traer entero: `git -C <repo> checkout 777e5f6a -- docs/ARCHITECTURE.md`.
   - Si develop SÍ lo cambió, usar `git restore -p --source=777e5f6a -- docs/ARCHITECTURE.md` y aceptar sólo los hunks de secciones nuevas (layout por módulo, reglas duras, observabilidad, lifecycle, errores, seguridad). Conservar lo que develop ya tenga.

4. **`README.md` y `docs/README.md` (modificados, COORDINAR con 019/021)**:
   - Estos diffs contienen renames de tooling (`staging-db-2-local-db`→`reset-local-db-from-prod`, `core/*`→`platform/*`, `ponti-frontend`→`web`, `migrate-up`→`db-migrate-up`). Son territorio de 019 (db scripts) / 021 (deploy) / 001 (platform).
   - Usar `git restore -p --source=777e5f6a -- README.md docs/README.md` y decidir hunk por hunk:
     - Si 019/021 ya van a renombrar esos comandos en sus PRs → **descartar esos hunks aquí** (evitar doble fuente).
     - Si 024 va primero y esos renames no entran en otra feature → traerlos.
   - Si hay duda, dejar README/docs.README SIN tocar en 024 (son cambios de tooling, no de doc-024 propiamente) y anotarlo en el PR. El núcleo de 024 son los 17 nuevos + ARCHITECTURE.md.

5. **Verificar**:
   - `git -C <repo> diff --check` (whitespace; trivial en `.md`).
   - `git -C <repo> diff 777e5f6a -- <paths>` debe ser vacío para los 17 nuevos (idénticos al SOURCE).
   - Revisar enlaces relativos en los `.md` (rutas tipo `docs/...`, `../CLAUDE.md`) — son referencias, no se validan en build, pero confirmar que apuntan a archivos del propio paquete o a código existente en develop.

6. **Commit + push + PR** (sólo cuando el humano lo pida):
   - `git -C <repo> add -A`
   - commit con el mensaje del título; push; `gh pr create --base develop`.

## Archivos enteros vs parciales

- **Enteros (whole-file)**: los 17 nuevos. `docs/ARCHITECTURE.md` puede ir entero si no hay conflicto.
- **Parciales (partial-hunks / manual-port)**: `README.md`, `docs/README.md` (y `ARCHITECTURE.md` si conflicto).

## Migraciones / tests a incluir

Ninguna. No hay migraciones ni tests en este flist.

## Dependencias previas

Ninguna feature debe ir antes a nivel de bloqueo. Las referencias de los docs (`Makefile:openapi`, `migrations_v4/000233`, `internal/data-integrity/handler.go`, `internal/shared/handlers/errors.go`, `cmd/api/main.go`) ya están en develop — verificado con `git cat-file -e`/`git grep`. Por eso 024 puede mergear de forma independiente.

## Coordinación con el otro repo (FE feature-024)

- **Orden recomendado**: **BE-first** para OpenAPI (publica `docs/openapi/swagger.yaml`), luego FE corre `yarn codegen:openapi`. Pero como el spec es piloto (2 endpoints), el orden no es crítico. Para la parte de docs puros (CRUDAR_PLAN, etc.) no hay dependencia de orden.
- El FE feature-024 trae su propia doc (`docs/`, `docs/audit`, `RESPONSIVE_GUIDELINES`, `PR-92.md`); no comparte archivos con este paquete.

## Comandos git SUGERIDOS (para un humano; NO ejecutar desde el agente)

```bash
REPO=/home/pablocristo/Proyectos/pablo/ponti/core
git -C "$REPO" checkout develop
git -C "$REPO" checkout -b pr/feature-024-openapi-and-docs-be

# 17 nuevos (enteros)
git -C "$REPO" checkout 777e5f6a -- CLAUDE.md CRUDAR_PLAN.md \
  docs/OPENAPI.md docs/ERROR_CATALOG.md docs/OBSERVABILITY.md \
  docs/crudar-lifecycle.md docs/archive-restore-policy.md docs/entity-capabilities.md \
  docs/customers-projects-lifecycle.md docs/DATA_INTEGRITY_CONTRACT.md \
  docs/MULTI_TENANT_100_EVIDENCE.md docs/BACKEND_CLEANUP_AUDIT.md \
  docs/audit-custom-errors.md docs/projects-archive-audit.md \
  docs/openapi/openapi.yaml docs/openapi/swagger.yaml docs/openapi/swagger.json

# ARCHITECTURE.md (entero si no hay conflicto, si no parcial)
git -C "$REPO" checkout 777e5f6a -- docs/ARCHITECTURE.md   # o: git restore -p --source=777e5f6a -- docs/ARCHITECTURE.md

# README / docs/README (parcial, coordinar con 019/021)
git -C "$REPO" restore -p --source=777e5f6a -- README.md docs/README.md

git -C "$REPO" diff --check
git -C "$REPO" status
```

## Qué NO traer

- `docs/openapi/docs.go` (no está en flist ni en SOURCE; regenerar con `make openapi`).
- Cambios de `Makefile`, `go.mod`, `go.sum`, código en `internal/**` o migraciones — fuera del flist.
- Los hunks de README/docs.README que ya estén cubiertos por 019/021 (evitar doble fuente / conflictos).

## Qué podría romperse

- Nada en runtime (docs). El único riesgo real es de **merge conflict** en los 3 modificados si develop diverge — se resuelve manual.
- Si por error se trae `docs/ARCHITECTURE.md` entero y develop ya había evolucionado el TLDR, se pierde el cambio de develop → revisar el diff antes de commitear.

## Cómo detectar extracción incompleta

- `git -C <repo> diff 777e5f6a -- <cada path nuevo>` debe ser vacío.
- `ls docs/openapi/` debe mostrar `openapi.yaml swagger.yaml swagger.json` (3, no `docs.go`).
- Enlaces dentro de los `.md` que apunten a docs hermanos del paquete (todos presentes).

## Qué validar antes del PR

- Que los 17 nuevos existen y son byte-idénticos al SOURCE.
- Que `README.md`/`docs/README.md` no revirtieron renames que develop ya tiene (si 019/021 fueron primero).
- `git diff --check` limpio.

## Qué hacer después de mergear

- Avisar al FE feature-024 que `docs/openapi/swagger.yaml` está disponible para `yarn codegen:openapi`.
- Backlog: anotar los ~48 handlers restantes (queda fuera de 024) y regenerar el spec; actualizar conteos en `MULTI_TENANT_100_EVIDENCE.md`/`BACKEND_CLEANUP_AUDIT.md` si se quieren mantener vigentes.
