# extraction-plan.md — feature-002 · be-crudar-lifecycle-framework

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip vacío/restore).
- **rama sugerida**: `pr/feature-002-be-crudar-lifecycle-framework-be`
- **merge**: BE-first. Sin contraparte FE. Bloquea feature 009.

## PR title

`feat(be): CRUDAR lifecycle framework (internal/shared/lifecycle + migraciones 227/228/232/233)`

## PR description (borrador)

> Introduce `internal/shared/lifecycle`: framework declarativo de archivado/
> restore (CRUDAR). Aporta tabla `archive_batches`, columnas de causa propagadas
> a ~32 tablas, mapa `Policies` por entidad, cascade archive/restore con scope
> por Cause, barrera de invariante "archived = no existe" (Go + restore-check +
> triggers Postgres) y métrica `crudar_rejected_archived_ref_total`.
>
> Solo agrega el paquete compartido y sus migraciones. Los consumidores
> (repositorios por entidad, endpoints archive/restore, CLI archive-cleanup)
> llegan en la feature 009 (crudar-archive-surface). Sin contraparte FE.
>
> PREREQUISITO: go.mod debe incluir `prometheus/client_golang`,
> `platform/observability/go` y `platform/persistence/gorm/go`. Migraciones
> 227/228 quedan numeradas por debajo de las ya mergeadas 229/230 — ver nota de
> ordenamiento.

## Pasos ordenados

0. **PREREQUISITO — dependencias.** Verificar/garantizar que `go.mod`/`go.sum`
   de develop tengan: `github.com/prometheus/client_golang v1.23.2` (+ indirects
   client_model/common/procfs), `github.com/devpablocristo/platform/observability/go v0.2.1`,
   `github.com/devpablocristo/platform/persistence/gorm/go v0.1.0`,
   `gorm.io/driver/sqlite v1.6.0` (este SÍ está en develop). Si faltan,
   coordinar con feature 021 o agregarlas en este PR como bloque previo. Sin
   esto el paquete NO compila.

1. Crear rama desde develop.
2. Traer los 9 archivos del paquete `internal/shared/lifecycle/` (whole-file).
3. Traer las 8 migraciones (4 pares) 227/228/232/233 (whole-file).
4. **Resolver ordenamiento de migraciones** (ver "Qué podría romperse").
5. `go mod tidy` (si se tocó go.mod), `go build ./internal/shared/lifecycle/...`,
   `go test ./internal/shared/lifecycle/...`, `go vet`.
6. `git diff --check` (trailing whitespace / conflict markers).
7. Validación de migraciones contra una DB de staging (up + down).

## Archivos enteros vs parciales

- **Enteros (whole-file)**: los 9 `.go` y las 8 `.sql`. Todos son `A` (no
  existen en develop), no hay hunks que separar.
- **Parciales (partial-hunks)**: `go.mod`/`go.sum` — SOLO si se decide agregar
  las deps en este PR. Hay que tomar ÚNICAMENTE las líneas de prometheus +
  observability/go + persistence/gorm/go, NO el resto del diff de go.mod del
  rango (que incluye cambios de otras features y los bumps go-jose/x/net YA
  porteados en #124 — excluir esos).

## Migraciones / tests a incluir

- Migraciones: 227, 228, 232, 233 (up+down). NO incluir 223/224/225/226/231/234
  (features 001/004).
- Tests: los 4 `*_test.go` del paquete. Corren con sqlite in-memory; no
  requieren Postgres ni fixtures externos.

## Dependencias previas

- go.mod/go.sum con las 3 deps nuevas (bloqueante).
- Las migraciones 232/233 asumen que existen en la DB: vistas
  `v4_report.workorder_list`, `v4_report.labor_list`; funciones
  `v4_ssot.labor_cost_for_lot`, `..._pre_harvest_for_lot`, `seeded_area_for_lot`,
  `supply_cost_for_lot` (232) y las tablas con `deleted_at` + FK (233). En una DB
  al día con develop deberían existir; validar.

## Coordinación con el otro repo

- **Solo-BE**. No hay PR de FE asociado. En el cross-repo-map del FE marcar
  "feature-002: sin cambios FE".

## Comandos git SUGERIDOS (para un humano; el agente NO los ejecuta)

```bash
# 1) rama
git checkout develop
git checkout -b pr/feature-002-be-crudar-lifecycle-framework-be

# 2) paquete Go (whole-file)
git checkout develop-problematico~1 -- internal/shared/lifecycle/

# 3) migraciones (whole-file)
git checkout develop-problematico~1 -- \
  migrations_v4/000227_crudar_archive_batches.up.sql \
  migrations_v4/000227_crudar_archive_batches.down.sql \
  migrations_v4/000228_crudar_remaining_archive_metadata.up.sql \
  migrations_v4/000228_crudar_remaining_archive_metadata.down.sql \
  migrations_v4/000232_document_archived_inclusion_in_reports.up.sql \
  migrations_v4/000232_document_archived_inclusion_in_reports.down.sql \
  migrations_v4/000233_archived_invariant_triggers.up.sql \
  migrations_v4/000233_archived_invariant_triggers.down.sql

# 4) deps (SOLO si NO vinieron de feature 021): editar a mano go.mod o usar
#    restore -p para tomar solo los hunks de prometheus/observability/persistence,
#    EXCLUYENDO los bumps go-jose/x/net (ya en #124).
git restore -p --source=develop-problematico~1 -- go.mod go.sum
go mod tidy

# 5) verificación
git diff --check
go build ./internal/shared/lifecycle/...
go test ./internal/shared/lifecycle/...
go vet ./internal/shared/lifecycle/...
```

## Qué NO traer

- 223/224/225/226/231/234 (otras features).
- `internal/*/repository.go`, `cmd/api/main.go`, `cmd/archive-cleanup/main.go`
  (features 009/010/023/019). El paquete debe compilar SIN ellos.
- Bumps go-jose/x/net en go.mod (YA en #124 — excluir de los hunks).

## Qué podría romperse

1. **Compilación**: si las 3 deps no están en go.mod → `go build` falla con
   "missing go.sum entry" / "no required module provides package
   github.com/prometheus/client_golang...". BLOQUEANTE.
2. **Orden de migraciones**: develop ya tiene 000229 y 000230. Mis 227/228 son
   números MENORES. Con golang-migrate estricto, una DB que ya está en 230+ NO
   acepta aplicar 227/228 (gap hacia atrás). Opciones:
   (a) renumerar 227/228/232/233 a números > 234 (p.ej. 235..238) — ROMPE las
       referencias cruzadas: 232 menciona "migration 000196", 233 dice "fase
       9-10"; los nombres internos del archivo no afectan al runner pero hay que
       actualizar consumidores que esperen el número (revisar feature 009).
   (b) confirmar que el runner del proyecto aplica por orden de archivo/permite
       out-of-order (revisar `scripts/` de migración y la memoria
       `reset-local-db-from-prod-old-migrations`).
   Recomendado: validar el runner antes de renumerar. NO renumerar a ciegas.
3. **232 COMMENT**: falla si una vista/función comentada no existe en la DB
   destino.
4. **233 triggers**: `CREATE TRIGGER` falla si alguna tabla referenciada no
   existe o no tiene `deleted_at`. La propia migración advierte: correr antes
   `scripts/data-audit/archived_invariants.sql` en staging (ese script NO está
   en mi flist; viene de feature 009/019).

## Cómo detectar extracción incompleta

- `go build ./...` del repo entero falla por símbolos `lifecycle.*` no resueltos
  → faltan archivos del paquete (improbable, son 9 y todos `A`).
- Si se trajeron consumidores por error (`internal/*/repository.go`) sin sus
  dependencias (actor-system feat 007/008), el build romperá por OTROS símbolos
  → señal de que se mezcló alcance.
- `git status` debe mostrar SOLO los 17 paths de esta feature (+ go.mod/go.sum
  si se decidió incluir deps).

## Qué validar antes del PR

- `go build` y `go test` del paquete en verde.
- `go build ./...` del repo entero en verde (el paquete no debe romper nada
  porque nadie lo importa todavía).
- Migraciones up/down sobre staging (idealmente sobre un dump de prod actualizado
  a develop). Ver validation.md.
- `git diff --check`.

## Qué hacer después de mergear

- Desbloquear feature 009 (crudar-archive-surface): traer los
  `internal/*/repository.go`, `cmd/archive-cleanup`, endpoints archive/restore y
  la llamada `lifecycle.RegisterMetrics` en `cmd/api/main.go`.
- Confirmar que el dashboard de Prometheus tiene namespace consistente
  (`ponti_backend`, ver `cmd/api/main.go` en 009).
