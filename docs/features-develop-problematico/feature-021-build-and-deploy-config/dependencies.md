# dependencies.md — feature-021 (BE)

## Depende de (intra-repo)

| feature | fuerza | qué aporta | por qué |
|---|---|---|---|
| 001 be-platform-tenancy-refactor | **fuerte** | `platform/persistence/gorm/go` usado por todos los `internal/*/repository.go` | sin ese código, agregar la dep a `go.mod` es prematuro y `go mod tidy` la borra |
| 005 be-config-modularization / 020 ci-workflows | **fuerte** | `platform/observability/go` (logger, metrics, tracer) en `cmd/api/main.go`, `http_server.go`, middlewares | la dep `observability/go` y el prefetch del Dockerfile dependen de este código |
| 002 be-crudar-lifecycle-framework | **fuerte** | `prometheus/client_golang` en `internal/shared/lifecycle/metrics.go` | la dep `prometheus/client_golang` viene de acá |
| 013 be-csv-export | **fuerte (negativa)** | define si `excelize/v2` se queda o se va | el target REMUEVE excelize; develop lo usa en `wire/*_providers.go` + `internal/platform/files/excel/`. No quitar excelize de `go.mod` hasta que 013 decida |
| 019 be-local-tooling-db-scripts | **débil** | `make db-schema-snapshot` genera `scripts/db/schema.snapshot.sql` | el único hunk de `.gitignore` portable hoy ignora ese archivo |
| 023 be-wire-di | **fuerte** | wire providers consumen platform/persistence | parte del mismo grafo de deps |

## Bloquea a (intra-repo)

- Ninguna feature de código depende de 021. Es config terminal. El cierre final de `go.mod`/`go.sum` se hace DESPUÉS de mergear las features de código (no antes), vía `go mod tidy`.

## Cross-repo

| dirección | feature | fuerza | nota |
|---|---|---|---|
| BE ↔ FE | 021 (espejo FE: vite/tailwind/eslint/knip/tsconfig/lockfiles/generated client) | **incierta/baja** | sin archivos compartidos. Sólo coordinación de naming/orden de PR. NO bloqueante |

## Archivos / tipos / config / migraciones / APIs compartidos

- `go.mod` / `go.sum`: **compartidos con TODO el backend**. Cambiarlos en 021 colisiona con cualquier feature de código que también los toque. Regla: 021 NO los toca; los regenera `go mod tidy` al final.
- `.gitignore`: compartido con 019 (entrada `schema.snapshot.sql`) y con tooling local (`go.work`).
- `Dockerfile`: compartido con 005/020 (prefetch de `observability`/`persistence`).
- APIs / migraciones: ninguna compartida por 021.

## Relación con lo YA DONE

- #124 (develop `3de0b453`) bumpea go-jose/x/net en `go.mod`/`go.sum`. **Excluir** esos hunks de cualquier port de 021.
- La migración `core→platform` de `go.mod`/`Dockerfile`/`docker-compose.yml` YA está en develop (commits `9a5e465b`, `e7bb89d8`, `93f77883`). 021 no la re-hace.

## Recomendación de orden

1. (ya hecho en develop) migración core→platform + #124.
2. Mergear features de código: 001, 002, 005/020, 013, 023.
3. `go mod tidy` → commitear `go.mod`/`go.sum` resultantes.
4. **Entonces** cerrar 021: hunk del Dockerfile prefetch + verificación final.
5. Los deltas hoy-portables de 021 (mount compose, ignore schema.snapshot) pueden ir **ya**, en paralelo, sin esperar nada.
6. FE-021: cuando quiera; sin dependencia técnica con BE-021.
