# file-list — feature-004 · shared-text / proper-name normalization

Fuente autoritativa: `/tmp/flists/be-004.txt`. Solo 2 archivos, ambos `A` (created), ambos propios y autocontenidos.

## Propios

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/shared/text/propername.go` | A | Go (paquete `text`) | Implementación: `CanonicalizeName` + `FormatProperName` + helpers privados (`stripDiacriticsPreservingEnye`, `formatWord`) y tablas `connectors`/`uppercaseTokens` | **whole-file** | Archivo nuevo, no existe en develop, sin hunks mezclados. Solo stdlib + `golang.org/x/text/unicode/norm` (ya en develop) | bajo | alta |
| `internal/shared/text/propername_test.go` | A | Go test (paquete `text`) | Tests table-driven de ambas funciones; fijan el contrato de comportamiento | **whole-file** | Archivo nuevo, self-contained, mismo paquete | bajo | alta |

## Compartidos (partial-hunks)

Ninguno. No hay archivos compartidos en esta feature. (No toca `wire/*`, `cmd/api/*`, `cmd/config/*`, `go.mod`, `go.sum`, `Makefile`, `internal/shared/handlers/**`, `internal/shared/models/base.go`, `internal/shared/repository/**`.)

## Requeridos por dependencia

Ninguno. La librería `golang.org/x/text` ya está en `go.mod` de develop (`v0.37.0`), por lo que NO hay que tocar go.mod/go.sum.

## Dudosos

Ninguno.

## NO traer todavía (callers fuera de feature-004)

Estos archivos consumen `internal/shared/text` en el source ref `777e5f6a` pero **NO** pertenecen a esta feature; se portan con 007/010/011/customer. No están en tu flist y NO deben extraerse aquí:

| path (referencia, NO extraer en 004) | feature dueña probable |
|---|---|
| `internal/actor/handler/dto/actor.go` | 007 actor-system |
| `internal/actor/master_link.go` | 007 actor-system |
| `internal/customer/handler/dto/requests.go` | customer (007/relacionada) |
| `internal/project/handler/dto/project.go` | 010 projects / 011 |
| `internal/project/repository.go` | 010 projects (compartido tenancy+dominio) |

Detectados con: `git -C <core> grep -ln "shared/text" 777e5f6a`.
