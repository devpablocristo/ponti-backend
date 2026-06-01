# extraction-plan.md — feature-023 · be-wire-di

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (tip = restore/vacío).
- **rama sugerida**: `pr/feature-023-be-wire-di-be`
- **naturaleza**: feature de costura (DI + bootstrap). **No es un PR autónomo**: sus hunks acompañan a los módulos. Este plan describe (a) el subconjunto verdaderamente propio de 023 y (b) cómo coordinar los hunks mezclados.

## PR title

`infra(be): wire DI graph & bootstrap (observability, slog migrate, archive-cleanup)`

## PR description (sugerida)

> Bootstrap de binarios y costura de Wire para la oleada new-cns3. Trae: logging estructurado slog + métricas Prometheus + tracing OTLP en `cmd/api`; migración de `cmd/migrate` a slog; nuevo binario `cmd/archive-cleanup` (dry-run/apply del cleanup de archive). Los hunks de `wire/*` que cablean actor/companion/data-integrity/CSV viajan con sus respectivas features (007/012/018/013) — este PR asume que ya están en develop.
>
> Depende de: #001, #002, #005, #007, #008, #009, #012, #013, #018.

## Estrategia de extracción

### Orden de features (esta NO va primera)
1. 001, 005 (platform + config) → 002, 008, 009 → 007, 012, 013, 018 → **023 al final**.
2. 023 cierra la costura: una vez que existen `actor.Handler`, `axis.CompanionClient`, `NewCSVExporter`, ports de data-integrity, y `lifecycle.RunArchiveCleanup`, recién ahí `wire_gen.go` compila.

### Archivos enteros (solo de 023)
- `cmd/archive-cleanup/main.go` (nuevo)
- `cmd/migrate/main.go`, `cmd/migrate/migrate_gorm.go`, `cmd/migrate/migrate_sql.go`, `cmd/migrate/migrate_sql_test.go`

### Archivos parciales (restore -p, MEZCLADOS)
- `cmd/api/main.go`, `cmd/api/http_server.go` — quedarte con hunks de observability/bootstrap; los hunks `ActorHandler` viajan con 007.
- `wire/wire.go`, `wire/wire_gen.go` — **no editar wire_gen.go a mano**; idealmente re-correr `go generate ./wire` (o `wire ./wire`) tras tener todos los providers. Si re-generás, `wire_gen.go` cae solo.
- `wire/ai_providers.go`, `wire/config_providers.go` → con 012.
- `wire/middleware_providers.go` → con 008.
- `wire/data_integrity_providers.go`, `wire/admin_providers.go` → con 018.
- `wire/lot|supply|stock|work_order|labor_providers.go` → con 013.

## Pasos ordenados

1. Asegurar 001/002/005/007/008/009/012/013/018 ya en `develop`.
2. `git checkout develop && git pull`
3. `git checkout -b pr/feature-023-be-wire-di-be`
4. Traer enteros los archivos propios:
   - `git checkout develop-problematico~1 -- cmd/archive-cleanup/main.go cmd/migrate/main.go cmd/migrate/migrate_gorm.go cmd/migrate/migrate_sql.go cmd/migrate/migrate_sql_test.go`
5. Traer parciales de bootstrap (solo hunks de observability):
   - `git restore -p --source=develop-problematico~1 -- cmd/api/main.go cmd/api/http_server.go`
6. Re-generar el grafo (preferido) en vez de copiar `wire_gen.go`:
   - `go run github.com/google/wire/cmd/wire ./wire` (o `go generate ./wire`)
   - Si NO podés regenerar, traer `wire/wire.go` y `wire/wire_gen.go` enteros desde SOURCE **solo cuando** todos los providers ya existan en develop.
7. Verificar costura: `go build ./...`
8. `go vet ./wire/... ./cmd/...`
9. `git diff --check` (whitespace).
10. `go test ./cmd/migrate/...`

## Migraciones / tests a incluir
- Migraciones: ninguna propia.
- Tests: `cmd/migrate/migrate_sql_test.go`.

## Dependencias previas (deben estar en develop antes del PR)
001, 002, 005, 007, 008, 009, 012, 013, 018. (Ver dependencies.md para fuerza de cada una.)

## Coordinación con el otro repo
- **BE-only**. No requiere coordinación con FE. Mencionar en cross-repo-map del FE: "feature-023 sin cambios FE".

## Comandos git SUGERIDOS (para un humano; NO ejecutar desde el agente)
```
git checkout develop
git checkout -b pr/feature-023-be-wire-di-be
git checkout develop-problematico~1 -- cmd/archive-cleanup/main.go cmd/migrate/main.go cmd/migrate/migrate_gorm.go cmd/migrate/migrate_sql.go cmd/migrate/migrate_sql_test.go
git restore -p --source=develop-problematico~1 -- cmd/api/main.go cmd/api/http_server.go
# providers mezclados: traer con su feature, o si todos los símbolos ya existen:
git restore -p --source=develop-problematico~1 -- wire/wire.go
go run github.com/google/wire/cmd/wire ./wire   # regenera wire_gen.go
git diff --check
go build ./... && go test ./cmd/migrate/...
```

## Qué NO traer
- `wire/actor_providers.go` (007) ni `wire/companion_providers.go` (012): no están en este flist.
- Hunks de CSV/data-integrity/companion en los providers si sus features aún no están en develop.

## Qué podría romperse
- `wire_gen.go` copiado a mano queda desfasado de `*_providers.go` → build roto difícil de leer.
- `axis.CompanionClient`, `NewCSVExporter`, `ActorSet`, `dataintegrity.SupplyRepositoryPort` inexistentes → `undefined:` en compilación.
- `MiddlewaresEnginePort.GetProtected()` removido: cualquier llamador que aún lo use rompe (verificar grep).

## Cómo detectar extracción incompleta
- `go build ./...` con errores `undefined: Provide...` o `undefined: axis.` → falta una feature dependiente.
- `wire: ... no provider found` al regenerar → falta un `*Set` o provider.

## Qué validar antes del PR
- `go build ./...`, `go vet ./...`, `go test ./cmd/migrate/...`.
- Arranque local: `go run ./cmd/api` con env mínimas (DB + `COMPANION_BASE_URL` + `COMPANION_INTERNAL_JWT_SECRET`).
- `curl localhost:8080/observability/metrics` → 200.
- `go run ./cmd/archive-cleanup --dry-run --output json`.

## Qué hacer después de mergear
- Regenerar `wire_gen.go` si quedó algo a mano (`go generate ./wire`).
- Confirmar que el deploy setea `OTEL_*` y `HTTP_RATE_LIMIT_PER_MINUTE` deseados (ver feature-021 build/deploy).
