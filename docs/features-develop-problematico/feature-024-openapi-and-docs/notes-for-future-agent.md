# notes-for-future-agent.md — feature-024 · openapi-and-docs (BE)

## Resumen corto

Paquete 100% documentación del backend. 20 entradas en el flist: 17 archivos nuevos (`.md`/`.yaml`/`.json`) + 3 docs modificados. CERO código, deps, migraciones o config. Es la feature más segura de extraer: básicamente `git checkout 777e5f6a -- <17 archivos>` para los nuevos. El único cuidado real está en los 3 archivos modificados (README, docs/README, ARCHITECTURE), que pueden conflictuar con 019/021/001.

## Qué está en FE y qué en BE (full-stack feature-024)

- **BE (este paquete)**: `CLAUDE.md`, `CRUDAR_PLAN.md`, `docs/OPENAPI.md`, catálogo de errores, observabilidad, lifecycle/archive, entity-capabilities, customers-projects-lifecycle, data-integrity-contract, multi-tenant-evidence, backend-cleanup-audit, audit-custom-errors, projects-archive-audit, y el spec OpenAPI (`docs/openapi/*`).
- **FE (otro repo, mismo feature-024)**: `docs/`, `docs/audit` (visual regression, posible generado), `RESPONSIVE_GUIDELINES`, `PR-92.md`. El FE consume `docs/openapi/swagger.yaml` con `yarn codegen:openapi`.
- No comparten archivos físicos; sólo el contrato OpenAPI (BE produce, FE consume).

## Archivos esenciales

- `docs/OPENAPI.md` + `docs/openapi/{openapi.yaml,swagger.yaml,swagger.json}`: el contrato. OJO: piloto de **2 endpoints** (`/me/context`, `/data-integrity/...`). NO es cobertura completa.
- `CLAUDE.md`, `docs/ARCHITECTURE.md`: guía de arquitectura hex + reglas duras.
- `docs/ERROR_CATALOG.md`, `docs/crudar-lifecycle.md`, `docs/archive-restore-policy.md`, `docs/entity-capabilities.md`: políticas canónicas.

## Archivos peligrosos / mezclados

- `README.md` y `docs/README.md` (MODIFICADOS): sus hunks renombran tooling (`staging-db-2-local-db`→`reset-local-db-from-prod`, `core/*`→`platform/*`, `ponti-frontend`→`web`, `migrate-up`→`db-migrate-up`). Esto pisa terreno de **019 (db scripts), 021 (deploy), 001 (platform)**. En develop el README AÚN dice los nombres viejos. NO traer estos hunks si 019/021 ya los aplicaron; preferir que esas features sean dueñas del rename.
- `docs/ARCHITECTURE.md` (MODIFICADO): el grueso es contenido nuevo seguro, pero el TLDR fue reescrito; verificar que develop no lo haya tocado antes de traerlo entero.
- `CRUDAR_PLAN.md`: es un plan **FE** (810 líneas) viviendo en el repo BE. Informativo, no normativo. No te confundas pensando que describe el estado del BE.
- Snapshots con datos viejos: `MULTI_TENANT_100_EVIDENCE.md`, `BACKEND_CLEANUP_AUDIT.md` (fechas 2026-05-12, `schema_migrations=225`, conteos). Son evidencia histórica, NO el estado actual.

## Decisiones ya tomadas

- 17 nuevos → whole-file (sin conflicto, no existen en develop).
- 3 modificados → partial-hunks / manual-port, coordinando con 019/021.
- NO crear `docs/openapi/docs.go` (no está en flist ni en SOURCE; se regenera con `make openapi`).
- 024 puede mergear INDEPENDIENTE; ninguna dep fuerte. Las referencias de los docs (Makefile:openapi, migrations_v4/000233, internal/data-integrity/handler.go) YA están en develop — verificado.

## Dudas abiertas (para humano)

1. ¿README.md/docs.README.md van con 024 o con 019/021? (overlap de tooling rename). Recomendación: que 019/021 sean dueños; 024 omite esos hunks si ya están.
2. ¿`CRUDAR_PLAN.md` debería estar en FE en vez de BE? El flist lo asigna a BE.
3. ¿Actualizar los snapshots de auditoría o conservarlos fechados?

## Qué comandos mirar primero

```bash
REPO=/home/pablocristo/Proyectos/pablo/ponti/core
cat /tmp/flists/be-024.txt
git -C "$REPO" diff --name-status 0972e565..777e5f6a -- CLAUDE.md CRUDAR_PLAN.md README.md docs/
git -C "$REPO" diff 0972e565..777e5f6a -- README.md docs/README.md docs/ARCHITECTURE.md   # ver los 3 modificados
git -C "$REPO" show 777e5f6a:docs/OPENAPI.md | head -60                                    # confirmar piloto 2 endpoints
git -C "$REPO" ls-tree 777e5f6a docs/openapi/                                              # confirmar 3 archivos, sin docs.go
# verificar que las referencias ya existen en develop:
git -C "$REPO" show 003a9b8f:Makefile | grep -n openapi
git -C "$REPO" ls-tree 003a9b8f migrations_v4/ | grep 000233
```

## Errores a evitar

- NO traer `docs/openapi/docs.go`.
- NO traer cambios de `Makefile`/`go.mod`/migraciones/`internal/**`: están fuera del flist (son de 019/020/021/002/...).
- NO revertir los renames de tooling que develop ya tenga al traer README/docs.README.
- NO usar `develop-problematico` (tip vacío); SOURCE es `777e5f6a` = `develop-problematico~1`.
- NO tratar los conteos de `MULTI_TENANT_100_EVIDENCE.md`/`BACKEND_CLEANUP_AUDIT.md` como estado actual.

## Camino más seguro

1. `git checkout -b pr/feature-024-openapi-and-docs-be` desde develop.
2. `git checkout 777e5f6a -- <17 nuevos>`.
3. `docs/ARCHITECTURE.md`: entero si no hay conflicto, si no `git restore -p`.
4. README/docs.README: omitir o `git restore -p` sólo hunks no cubiertos por 019/021.
5. `git diff 777e5f6a -- <17>` vacío + `git diff --check` limpio.
6. PR a develop, etiquetando el spec como piloto.

## PR del otro repo: antes/después

- Para la parte OpenAPI: **BE-024 antes** del `yarn codegen:openapi` del FE-024 (BE publica `swagger.yaml`). Para la doc pura, orden libre.
- El FE-024 trae su propia doc independiente; no bloquea ni es bloqueado por el contenido BE no-OpenAPI.
