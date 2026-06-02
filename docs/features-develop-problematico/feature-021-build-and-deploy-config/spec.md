# spec.md — feature-021 · Build & deploy config (Backend Go)

- **id:** feature-021
- **slug:** build-and-deploy-config
- **nombre:** Build & deploy config
- **tipo:** config
- **repo:** Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **merge:** por repo
- **existe-en-BE:** SÍ (este paquete)
- **existe-en-FE:** SÍ — FULL-STACK. El FE-021 cubre `vite` / `tailwind` / `eslint` / `knip` / `tsconfig` / lockfiles / generated client. NO se solapa en archivos con el BE.
- **SOURCE REF de extracción:** `develop-problematico~1` = SHA `777e5f6a` (NUNCA usar `develop-problematico`, su tip es un restore vacío).
- **rango fuente-de-verdad (diff):** `0972e565..777e5f6a`
- **rama destino:** `develop` (tip `003a9b8f`).

## Resumen

Cambios de configuración de build y despliegue del backend: `Dockerfile`, `docker-compose.yml`, `.gitignore`, y los archivos de módulos `go.mod` / `go.sum`. El objetivo nominal de la feature es dejar la imagen de build, el compose de dev y los ignores consistentes con la migración `core/* → platform/*` (parte de new-cns3) y con el tooling local actual.

## Objetivo

Que el backend buildee y corra (Docker + compose dev) consumiendo `github.com/devpablocristo/platform/*` en vez de `github.com/devpablocristo/core/*`, sin `replace` a paths locales, sin el mount histórico de `core/`, y con los `.gitignore` alineados al tooling de DB local nuevo.

## Problema

En `0972e565` el build dependía de `core/*` (deprecado), con un mount en `docker-compose.yml` a una ruta local fija (`/home/pablo/Projects/Pablo/core`) y `replace` directives. Eso rompe en CI/otros devs y atan el build a una topología de filesystem concreta.

## Alcance EN ESTE REPO (BE)

Mi flist (autoritativa, `/tmp/flists/be-021.txt`) son **5 archivos**:

| archivo | naturaleza del cambio en `0972e565..777e5f6a` |
|---|---|
| `Dockerfile` | prefetch de módulos `core_modules → platform_modules`; agrega `observability/go` y `persistence/gorm/go` a la lista de prefetch; borra línea en blanco final |
| `docker-compose.yml` | quita el mount `${CORE_REPO_DIR}:/home/pablo/Projects/Pablo/core`; quita `GOWORK=off`; agrega comentario NOTA explicando por qué |
| `.gitignore` | quita entradas de `scripts/db/*.env` y bloque `.env.*` comentado; quita bloque `go.work` / `go.work.sum` / `/api`; agrega ignore de `scripts/db/schema.snapshot.sql` |
| `go.mod` | **COMPARTIDO** — migración `core→platform` (YA en develop), + agrega `observability/go`, `persistence/gorm/go`, `prometheus/client_golang`, otel exporters; **quita** `excelize/v2` y `cloudsqlconn` y sus indirectas; bumps de deps |
| `go.sum` | **COMPARTIDO** — checksums derivados de `go.mod` (679 líneas de diff) |

### Hallazgo central (CRÍTICO)

`develop` y `develop-problematico` **divergieron desde el mismo merge-base `0972e565`** (verificado: `merge-base develop 777e5f6a == 0972e565`; `777e5f6a` NO es ancestro de develop). Buena parte de la intención de 021 **ya está en develop por otra vía**:

- `go.mod`/`Dockerfile`/`docker-compose.yml` de develop **ya están en `platform/*`** (commits develop `9a5e465b` "Dockerfile prefetch usa platform/*", `e7bb89d8` "backend pre-new-cns3 sobre platform + tooling local", `93f77883` "eliminar replace directives a paths locales").
- Lo que falta en develop respecto del target NO es "config de build" pura, sino **la consecuencia en el grafo de dependencias de features de código que todavía no se portaron**:
  - `platform/observability/go` → lo usan `cmd/api/main.go`, `cmd/api/http_server.go`, middlewares (`internal/platform/http/middlewares/gin/observability.go`), `internal/shared/handlers/errors.go` → **features 005/020/001**.
  - `platform/persistence/gorm/go` → lo usan **TODOS** los `internal/*/repository.go` → **feature 001 (platform-tenancy) / 023 (wire-di)**.
  - `prometheus/client_golang` → `internal/shared/lifecycle/metrics.go` → **feature 002 (crudar-lifecycle)**.
  - **Remoción** de `excelize/v2` (+ `xuri/*`, `richardlehane/*`, `tiendc/go-deepcopy`) → el target ya NO tiene `internal/platform/files/excel/...`; develop SÍ lo tiene y lo usa en `wire/*_providers.go` → **feature 013 (csv-export)**.
  - **Remoción** de `cloudsqlconn` → develop lo usa en `internal/platform/persistence/gorm/repository.go`.

Conclusión: el delta de `go.mod`/`go.sum` **no es portable como "config"**; se regenera solo con `go mod tidy` cuando aterricen las features de código que lo motivan.

## Alcance EN EL OTRO REPO (FE)

FE-021 (paquete espejo): `vite.config`, `tailwind`, `eslint`, `knip`, `tsconfig`, lockfiles (`yarn.lock`/`package-lock`), generated API client. Sin overlap de archivos con BE. La coordinación es de orden de merge, no de contenido (ver `dependencies.md`).

## Fuera de alcance (NO extraer aquí)

- **Dependency bumps `go-jose/v4` y `x/net`** → YA DONE en BE #124 (commit develop `3de0b453`). Excluir de 021.
- Cualquier `.go` (handlers, repositories, main, middlewares, wire). Pertenecen a 001/002/005/013/020/023.
- `Makefile`, `scripts/db/**` → feature 019.
- Limpieza de json-tags del dominio → feature 027.

## Comportamiento esperado

`docker compose up` levanta el backend resolviendo `platform/*` desde el proxy de módulos (sin mount local, sin replace). `docker build` prefetchea todos los módulos `platform/*` necesarios. El repo no trackea `schema.snapshot.sql` (generado por `make db-schema-snapshot`, feature 019).

## Estado en dp~1 (`777e5f6a`)

Config consistente con un backend que YA tiene todo el código de new-cns3 (observability, persistence/gorm, lifecycle metrics, sin excelize). En `develop` ese código NO está completo, por lo que la config de build de dp~1 **no aplica tal cual**.

## Criterios de aceptación

1. `Dockerfile`, `docker-compose.yml`, `.gitignore` quedan consistentes con el estado real de develop (platform, sin replace, sin mount local).
2. `go build ./...` y `go test ./...` compilan en develop.
3. `docker compose build && docker compose up` levanta el servicio.
4. **NO** se introducen deps (`observability`, `persistence/gorm`, `prometheus`, otel exporters) sin el código que las usa, ni se quitan deps (`excelize`, `cloudsqlconn`) cuyo código sigue presente en develop.

## Endpoints / modelos / UI / DB / tests afectados

- Endpoints: ninguno directo. (El target expone `GET /observability/metrics` pero eso es feature 005/020, no 021.)
- Modelos/DTOs: ninguno.
- UI: ninguna (BE).
- DB/migraciones: ninguna; sólo el ignore de `schema.snapshot.sql`.
- Tests: ninguno propio.

## Dependencias

- **Intra-repo:** depende fuerte de 001/002/005/013/020/023 para que el delta de `go.mod`/`go.sum` tenga sentido. Débil de 019 (ignore de `schema.snapshot.sql`).
- **Cross-repo:** FE-021 es independiente en archivos; sólo coordinación de orden.

## Riesgos

- **Funcional:** levantar la config de build de dp~1 contra un develop incompleto rompe el build (`go.mod` pediría deps que ningún `.go` usa → `go mod tidy` las borraría; o faltarían deps que el código de develop sí necesita: excelize).
- **Técnico:** `go.sum` es ruidoso (679 líneas), fácil de portar de más y romper la reproducibilidad.

## DECISIÓN recomendada

**PARTIR EN SUBFEATURES + POSTERGAR la parte de módulos.**

- **Extraer ahora (whole-file con cuidado / partial):** sólo los hunks de `Dockerfile`, `docker-compose.yml`, `.gitignore` que NO dependan de código no portado. En la práctica:
  - `docker-compose.yml`: quitar mount `core` + `GOWORK=off`, agregar NOTA. **Portable hoy.**
  - `.gitignore`: agregar ignore de `schema.snapshot.sql` (coordinar con 019). El resto del diff de `.gitignore` (quitar `go.work`, `scripts/db/*.env`) revierte tooling local que develop sí usa → **NO portar**.
  - `Dockerfile`: agregar `observability/go` + `persistence/gorm/go` al prefetch SÓLO cuando 001/005 aterricen (si no, prefetchea módulos que no se usan; no rompe pero es prematuro).
- **NO extraer `go.mod`/`go.sum` como parte de 021.** Dejar que se regeneren vía `go mod tidy` al portar las features de código. Excluir explícitamente go-jose/x/net (#124).
