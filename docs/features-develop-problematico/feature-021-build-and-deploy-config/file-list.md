# file-list.md — feature-021 (BE)

Flist autoritativa: `/tmp/flists/be-021.txt` (5 paths). Diff base `0972e565..777e5f6a`.

Leyenda extracción: `whole-file` / `partial-hunks` / `manual-port` / `do-not-extract-yet`.

## Propios de la feature (config de build/deploy)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `docker-compose.yml` | M | config compose | dev runtime; quita mount local `core` y `GOWORK=off`, agrega NOTA | **partial-hunks** | quitar el mount `${CORE_REPO_DIR}:/home/pablo/Projects/Pablo/core` y `GOWORK=off` es portable hoy; conservar el resto del block develop | bajo | alta |
| `.gitignore` | M | config | ignores; agrega `scripts/db/schema.snapshot.sql` | **partial-hunks** | SÓLO portar el hunk de `schema.snapshot.sql` (coord 019). NO portar el revert de `go.work`/`/api`/`scripts/db/*.env` que develop usa | medio | alta |
| `Dockerfile` | M | config build | prefetch de módulos `platform/*` | **manual-port** | develop YA está en platform; el único delta real es agregar `observability/go` + `persistence/gorm/go` al prefetch → portar SÓLO cuando 001/005 aterricen | bajo | alta |

## Compartidos (partial / NO de esta feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `go.mod` | M | módulos Go | grafo de deps | **do-not-extract-yet** | el delta vs develop NO es config: agrega deps (`observability`, `persistence/gorm`, `prometheus`, otel exporters) que sólo existen por código de 001/002/005, y quita `excelize`/`cloudsqlconn` cuyo código develop aún usa. Se regenera con `go mod tidy` | **alto** | alta |
| `go.sum` | M | checksums Go | reproducibilidad | **do-not-extract-yet** | 679 líneas derivadas de `go.mod`; nunca portar a mano. Excluye además los bumps go-jose/x/net (DONE #124) | **alto** | alta |

## Requeridos por dependencia

Ninguno dentro de mi flist. Las features que motivan el delta de `go.mod`/`go.sum` (001, 002, 005, 013, 020, 023) viven en otros paquetes.

## Dudosos

- `Dockerfile` línea del prefetch: si se porta el agregado de `observability/go`+`persistence/gorm/go` antes que el código que los usa, no rompe el build (sólo prefetchea de más), pero es prematuro y confunde. Tratado como `manual-port` condicionado.

## NO traer todavía (resumen)

- `go.mod`, `go.sum` (regenerar con `go mod tidy` al portar features de código).
- Hunks de `.gitignore` que revierten `go.work` / `go.work.sum` / `/api` / `scripts/db/*.env` (tooling local vigente en develop, feature 019).
- Bumps go-jose / x/net en `go.mod`/`go.sum` (DONE #124).
