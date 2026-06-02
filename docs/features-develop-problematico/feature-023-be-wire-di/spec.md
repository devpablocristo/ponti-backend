# spec.md — feature-023 · be-wire-di

- **id**: feature-023
- **slug**: be-wire-di
- **nombre**: Wire DI graph & bootstrap
- **tipo**: infra
- **repo**: Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe-en-FE**: NO. Solo-BE. En el cross-repo-map del FE se menciona como "sin cambios FE".
- **existe-en-BE**: SI.
- **merge**: acompaña a su módulo (no se mergea sola; sus hunks viajan junto a los módulos que cablea).

## Resumen

Esta feature es el **grafo de inyección de dependencias (Google Wire)** y el **bootstrap de los binarios** (`cmd/api`, `cmd/migrate` y el nuevo `cmd/archive-cleanup`). No tiene lógica de negocio propia: es el "pegamento" que instancia repos, use-cases y handlers de cada módulo y los registra en el router Gin.

El diff `0972e565..777e5f6a` mezcla en estos archivos cambios que pertenecen a varias features:
- Alta del `ActorHandler`/`ActorSet` (→ feature-007).
- Migración del cliente AI de `ai.Client` a `axis.CompanionClient` + adapter, y alta de `ProvideNexusClient` (→ feature-012).
- Re-cableado del módulo `data-integrity` (cambia firmas de `ProvideDataIntegrityUseCases` y pasa repos concretos en vez de ports; agrega supply/project, saca stock) (→ feature-018).
- Reemplazo de los exporters Excel (`*ExcelService`/`XLSXEnginePort`) por exporters CSV (`NewCSVExporter`) en lot/supply/stock/work-order/labor (→ feature-013).
- Observability: `slog` JSON logger, `observability.Metrics`, `TracerProvider` OTLP, middleware `ObservabilityWithMetrics`, endpoint `/observability/metrics`, CORS + rate-limit (→ infra propia de 023, acompaña a 005/008).
- Nuevo binario `cmd/archive-cleanup` que consume `internal/shared/lifecycle.RunArchiveCleanup` (→ consume framework de feature-002 / superficie de archive 009).

## Objetivo

Mantener el grafo de Wire compilable y coherente tras la oleada de refactors (actor, companion, data-integrity, CSV exporters, observability), y bootstrapear los binarios con logging estructurado, métricas, tracing y los middlewares globales nuevos (CORS, rate-limit, observability).

## Problema

Los archivos de wiring son **puntos de costura obligados**: cada feature que agrega un módulo o cambia una firma de constructor edita `wire/*_providers.go`, `wire/wire.go` y el archivo generado `wire/wire_gen.go`. Por eso el diff de esta feature es intrínsecamente **MEZCLADO** y no se puede portear "tal cual" sin arrastrar features que todavía no estén mergeadas (007, 012, 013, 018). Extraer 023 de forma aislada **no compila**.

## Alcance en este repo (BE)

- `cmd/api/main.go` — bootstrap: slog, metrics, tracer OTLP, anotaciones swagger, `os.Exit` en vez de `log.Fatalf`, nueva firma `runHTTPServer(ctx, logger, metrics, deps)`.
- `cmd/api/http_server.go` — middlewares globales nuevos (Observability+Metrics, CORS, RateLimit), endpoint `/observability/metrics`, registro de `deps.ActorHandler.Routes()`, sección `reporting.read_mode` en healthcheck.
- `cmd/archive-cleanup/main.go` — **binario nuevo**: CLI con flags `--apply/--dry-run/--tenant-id/--output`, invoca `lifecycle.RunArchiveCleanup`, imprime reporte tabla/json.
- `cmd/migrate/*` — migración de `log` a `slog`; firmas de `runMigrations`/`runGormMigrations`/`acquireMigrationLock` reciben `*slog.Logger`; test actualizado.
- `wire/wire.go` y `wire/wire_gen.go` — agregan `ActorSet`/`ActorHandler`, recablean AI→Companion+Nexus, data-integrity, admin (use-cases), exporters CSV.
- `wire/*_providers.go` (admin, ai, config, data_integrity, labor, lot, middleware, stock, supply, work_order) — providers por módulo.

## Alcance en el otro repo (FE)

Ninguno. Sin cambios en FE.

## Fuera de alcance

- `wire/actor_providers.go` (feature-007) y `wire/companion_providers.go` (feature-012): **NO** están en este flist; se traen con sus features.
- La lógica de `internal/shared/lifecycle/*` (feature-002) y la superficie de archive (feature-009): se asumen ya presentes.
- Implementaciones de `NewCSVExporter` en cada módulo (feature-013): se asumen presentes.
- Refactor del módulo data-integrity en sí (feature-018).

## Comportamiento esperado

- `go build ./...` y `go run cmd/api` arrancan el server con logger JSON, métricas Prometheus en `/observability/metrics`, tracing OTLP opcional (env `OTEL_*`), CORS y rate-limit configurables por env.
- `wire.Initialize()` construye todo el grafo sin ciclos ni providers faltantes.
- `cmd/archive-cleanup --dry-run` reporta sin mutar; `--apply` remedia; `--output json|table`.
- `cmd/migrate` (SQL y `-gorm`) loguea en JSON estructurado.

## Estado en dp~1 (SHA 777e5f6a)

Compila y funciona **dentro de develop-problematico** porque ahí conviven todas las features (actor, companion, data-integrity, CSV, lifecycle). Aislado sobre `develop` actual, **no compila** sin sus dependencias.

## Criterios de aceptación

1. `go build ./...` verde con 007, 012, 013, 018, 002 presentes.
2. `wire.Initialize()` no devuelve error en arranque.
3. `go test ./cmd/migrate/...` verde.
4. `cmd/archive-cleanup --dry-run --output json` produce JSON válido.
5. `/observability/metrics` responde 200; healthcheck incluye `reporting.read_mode`.

## Endpoints / rutas afectadas

- Nuevo: `GET /observability/metrics` (handler Prometheus, fuera de `/api/v1`).
- Registro de rutas del módulo actor: `deps.ActorHandler.Routes()` (rutas definidas en 007).
- Healthcheck (root): agrega `reporting.read_mode`.

## Modelos / DTOs / tipos

- `wire.Dependencies` agrega campo `ActorHandler *actor.Handler`.
- `MiddlewaresEnginePort` elimina el método `GetProtected()`.
- AI: tipos `config.Companion`, `axis.CompanionClient`, `ai.NewCompanionAdapter` (definidos en 012).
- data-integrity: `dataintegrity.SupplyRepositoryPort`, `ProjectRepositoryPort` (definidos en 018); deja de usar `StockRepositoryPort`.

## UI

N/A (solo BE).

## DB / migraciones

Ninguna migración propia. `cmd/archive-cleanup` y `cmd/migrate` operan sobre el esquema; las migraciones de archive viven en feature-002 (`migrations_v4/000227..000233`).

## Tests afectados

- `cmd/migrate/migrate_sql_test.go` (M) — pasa `*slog.Logger` a `acquireMigrationLock`.

## Dependencias

### Intra-repo (fuertes)
- **001** be-platform-tenancy-refactor — `platform/*` (observability, http/gin) que importan main/http_server.
- **005** be-config-modularization — `config.Companion`, `config.Reporting`, `config.Security.AITenantScope`, `cfg.Auth.RequireTenantHeader`, `cfg.Service.Env`.
- **007** actor-system — `internal/actor`, `wire/actor_providers.go` (`ActorSet`, `ProvideActor*`).
- **008** identity-tenant-context — `RequireTenantHeader`, `Environment` en middleware auth.
- **009** crudar-archive-surface — superficie de borrado/archive que `archive-cleanup` recorre.
- **012** ai-companion-integration — `axis.CompanionClient`, `companion_providers.go`, `ai.NewCompanionAdapter`, `NexusClient`.
- **002** be-crudar-lifecycle-framework — `internal/shared/lifecycle.RunArchiveCleanup`, `RegisterMetrics`.
- **013** be-csv-export — `NewCSVExporter` en lot/supply/stock/work-order/labor.
- **018** data-integrity-admin — firmas nuevas de `ProvideDataIntegrityUseCases` + ports supply/project.

### Cross-repo
- Ninguna. Solo-BE.

## Riesgos

### Funcionales
- Si falta `COMPANION_BASE_URL`/`COMPANION_INTERNAL_JWT_SECRET`, `ProvideCompanionClient` falla y el binario **no arranca** (cutover sin fallback, intencional).
- Rate-limit con `HTTP_RATE_LIMIT_PER_MINUTE` mal calibrado puede rechazar tráfico legítimo (default 0 = off).

### Técnicos
- `wire_gen.go` es **generado**: cualquier extracción parcial inconsistente con `wire.go`/`*_providers.go` rompe el build de forma sutil. No editar a mano sin re-correr `wire`.
- Mezcla de hunks: traer 023 sin 007/012/013/018 deja referencias a símbolos inexistentes.

## DECISIÓN recomendada

**Partir / portear acompañando a sus módulos (NO extraer como PR aislado).** Los hunks de `wire.go`/`wire_gen.go`/`cmd/api/main.go`/`http_server.go` son MEZCLADOS y deben viajar con `git restore -p` junto a cada módulo (actor→007, companion→012, data-integrity→018, CSV→013). Lo verdaderamente "propio de 023" (observability bootstrap, slog en migrate, binario `archive-cleanup`) se puede agrupar en un PR de infra **solo después** de que 001/002/005/008/009 estén en develop. Confianza: **alta** (basada en diff real y overlap de flists 007/012/002).
