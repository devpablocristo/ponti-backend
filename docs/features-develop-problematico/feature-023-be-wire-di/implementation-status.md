# implementation-status.md — feature-023 · be-wire-di

## Estado global
- **Estado**: completa **dentro de develop-problematico** (compila y arranca con todas las features presentes). **Rota si se extrae aislada** sobre develop actual (faltan símbolos de 007/012/013/018/002).
- **% completitud (como código en dp~1)**: ~100% funcional en su contexto.
- **% portabilidad aislada a develop**: ~30% (solo `cmd/migrate/*` y `cmd/archive-cleanup` son portables solos; el resto depende de otras features).

## Estado en este repo (BE)
- `cmd/migrate/*`: completo, autocontenido (solo logging slog). **Portable ya**.
- `cmd/archive-cleanup/main.go`: completo, pero requiere `internal/shared/lifecycle` (002) en develop.
- `cmd/api/main.go` + `http_server.go`: observability/bootstrap completo; el hunk `ActorHandler.Routes()` exige 007.
- `wire/*_providers.go` + `wire.go` + `wire_gen.go`: completos pero **mezclados** con 007/012/013/018.

## Estado en el otro repo (FE)
- No aplica. Sin cambios FE.

## Tests
- `cmd/migrate/migrate_sql_test.go`: actualizado para slog; debería pasar (`go test ./cmd/migrate/...`).
- No hay tests propios para wire ni para archive-cleanup CLI en este flist (la lógica testeada vive en `internal/shared/lifecycle/*_test.go`, feature-002).

## Pendientes / clasificación

### BLOQUEANTE para mergear
- Tener 001/002/005/007/008/009/012/013/018 en develop antes que 023; de lo contrario el build rompe.
- `wire_gen.go` debe ser coherente con los `*_providers.go` (regenerar con `wire`, no copiar a mano si hubo conflictos).
- Confirmar que `ProvideDataIntegrityUseCases` matchea EXACTO el orden de args de `dataintegrity.NewUseCases` de 018 (el diff cambió el orden: dashboard, workorder, report, supply, project, lot).

### Mejora futura
- Agregar test de humo de `wire.Initialize()` (hoy no hay).
- Documentar las env `OTEL_*` y `HTTP_RATE_LIMIT_PER_MINUTE` en el README de deploy (021).

### Deuda aceptable
- `ProvideNexusClient` se invoca y su resultado se descarta (`_`) "hasta ola 2": provider construido pero no usado. Aceptable, intencional.
- Comentarios de los providers ahora explican por qué data-integrity recibe repos concretos en vez de ports.

### Duda humana
- ¿El split admin (`ProvideAdminRepository`/`ProvideAdminUseCases`) pertenece a 018 o es refactor propio de 023? Resolver con `git log -- internal/admin` y flist be-018.

## Bugs observados
- Ninguno funcional en el diff. Riesgo principal es de **extracción** (símbolos faltantes), no de lógica.
