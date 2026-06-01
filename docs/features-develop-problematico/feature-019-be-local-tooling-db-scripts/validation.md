# validation.md — feature-019 · be-local-tooling-db-scripts

## Checklist pre-PR

- [ ] Rama creada desde `develop` (`pr/feature-019-be-local-tooling-db-scripts-be`).
- [ ] 14 archivos propios traídos enteros + `Makefile` por hunks.
- [ ] `smoke-companion/main.go` traído SOLO si `internal/axis` (012) está en `develop`.
- [ ] `repair_stocks_investor_granularity.sql` y `schema.snapshot.sql` borrados (`git rm`).
- [ ] `.gitignore` cubre `scripts/db/schema.snapshot.sql`.
- [ ] Bit +x en scripts ejecutables (`reset-local-db-from-prod.sh`, `lint-tenant-leaks.sh`, `export-ai-conversations.sh`, `run_ponti_local.sh`, `down_ponti_local.sh`).
- [ ] `git diff --check` limpio (sin trailing whitespace / conflict markers).
- [ ] `git diff develop..HEAD -- Makefile` NO contiene `openapi:`, `swag`, `golangci-lint v2.11.4`, ni renames `core/*`→`platform/*`.

## Validación automática (comandos)

```
# Sintaxis de shell
bash -n scripts/lint-tenant-leaks.sh
bash -n scripts/export-ai-conversations.sh
bash -n scripts/run_ponti_local.sh
bash -n scripts/down_ponti_local.sh
bash -n scripts/db/reset-local-db-from-prod.sh
bash -n scripts/db/db_migrate_up.sh
bash -n scripts/db/db_schema_diff.sh

# Make targets resuelven (dry-run, no ejecuta)
make -n reset-local-db-from-prod
make -n actors-backfill-sync
make -n up-ponti-local
make -n down-ponti-local

# Targets viejos eliminados NO deben existir
grep -n "db-staging-to-local\|db-reset-from-staging\|staging-db-2-dev-db\|db-force-reset-gcp\|db-gcp-reset-and-load-local" Makefile   # esperado: 0 hits

# Go: smoke-companion compila (solo si se incluyó)
go build ./scripts/smoke-companion

# Nada más del repo se rompe
go build ./...
```

## Tests sugeridos

- **BE:** no hay paquete de test propio. `go build ./scripts/smoke-companion` y `go build ./...`
  hacen de smoke de compilación. Opcional: `go vet ./scripts/smoke-companion`.
- **FE:** N/A (sin cambios FE).

## Manual / casos borde

1. `DRY_RUN=1 ./scripts/db/reset-local-db-from-prod.sh` → debe imprimir origen/destino y salir 0 sin tocar nada.
2. `reset-local-db-from-prod.sh` con `DB_HOST` distinto de localhost → debe abortar con "destino bloqueado".
3. `./scripts/lint-tenant-leaks.sh` → exit 0/1 sin error de sintaxis; revisar que escanee `internal/`.
4. `make actors-backfill-sync` sobre DB local restaurada → idempotente (correrlo 2 veces da igual resultado).
5. `psql ... -f scripts/db/tenant_isolation_audit.sql` → 0 filas en cada sección (DB con tenancy de 003).
6. `psql ... -f scripts/db/archived_invariants.sql` → cada check reporta count + sample ids.
7. `make up-ponti-local` sin `web/` presente → WARN (no debe romper el arranque de core).
8. `go run ./scripts/smoke-companion` con axis local arriba → ejercita `POST /v1/chat`.

## Qué revisar en UI / API / DB / env

- **UI:** nada.
- **API:** nada propio. `smoke-companion` consume `POST /v1/chat` de axis (externo).
- **DB:** que los `.sql` corran contra una DB con el schema de 003/007 aplicado.
- **env:** `.env` con `DB_*`; `SRC_*` / Secret Manager para PROD; `PONTI_AI_DSN` para export-ai;
  `PONTI_WEB_DIR` opcional para run/down.

## Qué validar en el otro repo

Nada. Sin cambios FE.

## Señales de incompletitud / incompatibilidad

- `make -n reset-local-db-from-prod` referencia un script que no existe → extracción incompleta.
- Reaparece `schema.snapshot.sql` en `git status` → no se borró.
- `go build ./scripts/smoke-companion` falla con "package internal/axis is not in std" → falta 012.
- El diff del Makefile contra develop muestra `openapi:`/`swag` → se trajeron hunks de más (024).
