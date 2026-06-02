# notes-for-future-agent.md — feature-019 · be-local-tooling-db-scripts

## Resumen corto

Tooling operativo puro: `scripts/**` + hunks del `Makefile`. Bajo riesgo. NO toca
código de la app, deps ni migraciones del schema. Solo-BE. Merge independiente.
19 paths en `/tmp/flists/be-019.txt` (14 propios enteros + 1 Makefile parcial + 2 borrados + 2 docs).

## Qué está en FE y en BE

- **BE:** todo. **FE (web):** nada → "sin cambios FE" en el cross-repo-map.
- El Makefile/`run_ponti_local.sh` *mencionan* `web/` y `axis/` por path, pero no editan esos repos.

## Archivos esenciales

- `scripts/db/reset-local-db-from-prod.sh` — el corazón del flujo de DB local (PROD→local data-only).
- `scripts/lint-tenant-leaks.sh` — guard de CI anti-fugas de tenancy (grep `internal/`).
- `scripts/db/{actors_backfill_sync,actors_golden_master,multi_tenant_golden_master,tenant_isolation_audit}.sql` — backfill + golden masters + auditoría.
- `scripts/README.md`, `scripts/data-audit/README.md` — índice y guía.

## Archivos peligrosos

- `scripts/db/reset-local-db-from-prod.sh` — destructivo en destino local; origen PROD read-only.
  Tiene guarda `DB_HOST` localhost y `DRY_RUN`. Correr siempre con `DRY_RUN=1` primero.
- `scripts/db/actors_backfill_sync.sql` — hace `TRUNCATE` de 4 tablas de actors antes de reinsertar (idempotente).

## Archivos mezclados (compartidos)

- **`Makefile`** — ÚNICO compartido. Su diff mezcla 019 con: 024 (`openapi:`/swag),
  001 (rename `core/*`→`platform/*`), 020/022 (`golangci-lint v2.11.4`), baja de
  `seed`/`select-ponti-*` (coordinar 005), `cmd/`→`cmd/api` (¿021/023?).
  EXTRAER POR HUNKS (`git restore -p`), no entero.

## Decisiones ya tomadas

- Extraer scripts enteros + Makefile por hunks.
- Borrar `repair_stocks_investor_granularity.sql` (one-shot ya aplicado) y `schema.snapshot.sql` (generado, 6589 líneas).
- `smoke-companion/main.go` = condicionado a feature-012 (`internal/axis`); si 012 no está, postergarlo.
- NO traer hunks del Makefile que pertenecen a otras features.

## Dudas abiertas

- ¿Qué feature posee el hunk `cmd/`→`cmd/api` del Makefile (021 o 023)?
- ¿La baja de `seed`/`seed-dashboard`/`select-ponti-*` va en 019 o en 005 (config)? Recomendación: NO en 019.

## Qué comandos mirar primero

```
cat /tmp/flists/be-019.txt
git -C <core> diff 0972e565..777e5f6a -- Makefile
git -C <core> diff 0972e565..777e5f6a -- scripts/db/reset-local-db-from-prod.sh
git -C <core> show 777e5f6a:scripts/README.md
git -C <core> show 777e5f6a:scripts/lint-tenant-leaks.sh
git -C <core> ls-tree -d 777e5f6a internal/axis   # confirmar dep de smoke-companion
```

## Errores a evitar

- Usar `develop-problematico` como fuente (tip = restore vacío). SIEMPRE `develop-problematico~1` / `777e5f6a`.
- Traer el Makefile entero (arrastra openapi/lint-v2/platform-rename).
- Incluir `smoke-companion/main.go` sin 012 → rompe `go build ./scripts/...`.
- Olvidar el `git rm` de los 2 archivos borrados.
- Ejecutar comandos git que muten (este flujo es solo lectura + Write de docs).

## Camino más seguro

1. Verificar que 012 (`internal/axis`) esté en `develop`. Si no → postergar smoke-companion.
2. `git checkout develop-problematico~1 -- <14 scripts propios>`.
3. `git restore -p --source=develop-problematico~1 -- Makefile` (solo hunks de tooling-DB).
4. `git rm` los 2 borrados.
5. Validar: `bash -n`, `make -n`, `go build ./...`, `git diff --check`.

## Qué PR del otro repo debe ir antes/después

Ninguno (solo-BE). No hay PR de web que coordinar para esta feature.
Internamente: idealmente 001/003/007/012 ya en `develop` (012 bloquea solo smoke-companion;
001/003/007 hacen útil el tooling pero no bloquean el merge).
