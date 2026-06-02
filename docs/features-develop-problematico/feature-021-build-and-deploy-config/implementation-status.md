# implementation-status.md — feature-021 (BE)

## Estado global

**PARCIALMENTE YA EN DEVELOP / RESTO NO-PORTABLE-TODAVÍA.**

La intención principal (core→platform en build/compose) ya aterrizó en develop por otra rama. El delta restante de dp~1 está **acoplado a código aún no portado** y no debe extraerse como "config".

- **% completitud de la intención 021 en develop:** ~70% (la migración platform de build/compose ya está; falta sólo el cierre de deps + 2 hunks de config menores).
- **% de mi flist que es extraíble HOY de forma segura:** ~25% (dos hunks: compose mount/GOWORK, gitignore schema.snapshot).

## Estado en este repo (BE / develop)

| archivo | en develop | delta vs target | acción |
|---|---|---|---|
| `Dockerfile` | platform_modules OK | falta `observability/go` + `persistence/gorm/go` en prefetch; sobra línea en blanco | diferir (manual-port cuando 001/005 estén) |
| `docker-compose.yml` | platform OK pero con mount `core` + `GOWORK=off` | quitar mount + GOWORK, agregar NOTA | **portable HOY** |
| `.gitignore` | tiene `go.work`/`/api`/`scripts/db/*.env` | agregar ignore `schema.snapshot.sql`; (target además revierte tooling local) | portar SÓLO el hunk schema.snapshot |
| `go.mod` | platform OK + excelize + cloudsqlconn | target agrega observability/persistence/prometheus/otel, quita excelize/cloudsqlconn, bumps | **NO portar** (regenerar con `go mod tidy`) |
| `go.sum` | derivado de go.mod develop | 679 líneas de diff | **NO portar** |

## Estado en el otro repo (FE)

Desconocido en detalle desde acá. FE-021 cubre vite/tailwind/eslint/knip/tsconfig/lockfiles/generated client. Sin overlap de archivos. Verificar en el paquete FE-021.

## Tests

- No hay tests propios de 021. Validación = `go build`/`go test` (de otras features) verdes + `docker compose build/up`.

## Pendientes

- Portar 2 hunks (compose, gitignore). **HOY.**
- Cerrar `go.mod`/`go.sum` vía `go mod tidy` tras mergear features de código. **DESPUÉS.**
- Agregar prefetch del Dockerfile. **DESPUÉS.**

## Clasificación de pendientes

### BLOQUEANTE para mergear (este PR mínimo)
- Asegurar que el PR NO toca `go.mod`/`go.sum`.
- Confirmar que el hunk de `.gitignore` no revierte `go.work`/tooling local de 019.

### Mejora futura
- Agregar `observability/go`+`persistence/gorm/go` al prefetch del `Dockerfile` (acelera build CI) una vez aterrizado el código.

### Deuda aceptable
- `go.mod` de develop carga `excelize`/`cloudsqlconn`: correcto mientras ese código exista. Se limpia con 013/`go mod tidy`.

### Duda humana
- ¿013 (csv-export) mantiene excelize o migra a CSV puro? De eso depende si `go.mod` finalmente conserva o no `excelize/v2`. Revisar paquete de feature-013.

## Bugs conocidos

- Ninguno en los archivos de config. El único "bug" potencial es de proceso: portar `go.mod`/`go.sum` whole-file rompe el build (ver risks.md).
