# file-list.md — feature-023 · be-wire-di

Fuente: `cat /tmp/flists/be-023.txt` (22 paths). Diff: `0972e565..777e5f6a`. SOURCE = `develop-problematico~1` (777e5f6a).

Leyenda extracción: `whole-file` = traer el archivo entero · `partial-hunks` = `git restore -p` (mezclado, solo algunos hunks son de 023) · `manual-port` = aplicar a mano · `do-not-extract-yet` = no traer hasta que su feature dueña esté en develop.

## Propios (infra 023)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `cmd/archive-cleanup/main.go` | A | binario CLI | Comando nuevo: dry-run/apply de cleanup de archive, output table/json | whole-file | Archivo nuevo, sin overlap. Consume `lifecycle.RunArchiveCleanup` (002) | medio (depende de 002) | alta |
| `cmd/migrate/main.go` | M | bootstrap | `log`→`slog`, `os.Exit`, pasa logger a runners | whole-file | Cambio autocontenido de logging | bajo | alta |
| `cmd/migrate/migrate_gorm.go` | M | bootstrap | `runGormMigrations(ctx, logger, repo)` con slog | whole-file | Cambio autocontenido | bajo | alta |
| `cmd/migrate/migrate_sql.go` | M | bootstrap | firmas reciben `*slog.Logger` en runners y lock | whole-file | Cambio autocontenido | bajo | alta |
| `cmd/migrate/migrate_sql_test.go` | M | test | actualiza llamada a `acquireMigrationLock` con logger | whole-file | Acompaña a migrate_sql.go | bajo | alta |

## Compartidos (partial-hunks — MEZCLADOS, traer con cada feature dueña)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `wire/wire.go` | M | DI manual | Declara `Dependencies` + `Initialize` con todos los `*Set` | partial-hunks | Hunks de 023 (observability nada acá) + ActorSet/ActorHandler (007) | alto | alta |
| `wire/wire_gen.go` | M | DI generado | Grafo materializado (NO editar a mano) | partial-hunks | Mezcla actor(007)+companion(012)+data-integrity(018)+CSV(013) | **alto** | alta |
| `cmd/api/main.go` | M | bootstrap | Entrypoint API: slog+metrics+tracer+swagger | partial-hunks | Observability (023) + cambio firma runHTTPServer | medio | alta |
| `cmd/api/http_server.go` | M | bootstrap | Router: middlewares globales, /observability/metrics, ActorHandler.Routes() | partial-hunks | Observability/CORS/rate-limit (023) + `ActorHandler.Routes()` (007) + `reporting.read_mode` (005/018) | alto | alta |
| `wire/ai_providers.go` | M | provider | AI: de `ai.Client` a `axis.CompanionClient`+adapter, +Nexus | partial-hunks | **Pertenece a 012** (companion). Tenant scope AI (008) | alto | alta |
| `wire/config_providers.go` | M | provider | `ProvideConfigAI`→`ProvideConfigCompanion` | partial-hunks | Acompaña a 012 + config 005 | medio | alta |
| `wire/middleware_providers.go` | M | provider | Quita `GetProtected()`; agrega `Environment`/`RequireTenantHeader` a auth | partial-hunks | Pertenece a 008 (identity/tenant) | medio | alta |
| `wire/data_integrity_providers.go` | M | provider | Recablea use-cases: +supply +project, -stock; repos concretos | partial-hunks | **Pertenece a 018** | alto | alta |
| `wire/admin_providers.go` | M | provider | Introduce `ProvideAdminRepository`/`ProvideAdminUseCases` | partial-hunks | Refactor admin (parte de 018/admin) | medio | media |
| `wire/lot_providers.go` | M | provider | Excel→CSV (`NewCSVExporter`), saca XLSX engine | partial-hunks | **Pertenece a 013** (CSV export) | medio | alta |
| `wire/supply_providers.go` | M | provider | Excel→CSV | partial-hunks | **Pertenece a 013** | medio | alta |
| `wire/stock_providers.go` | M | provider | Excel→CSV | partial-hunks | **Pertenece a 013** | medio | alta |
| `wire/work_order_providers.go` | M | provider | Excel→CSV (`ProvideWorkOrderExporterPort`) | partial-hunks | **Pertenece a 013** | medio | alta |
| `wire/labor_providers.go` | M | provider | Excel→CSV | partial-hunks | **Pertenece a 013** | medio | alta |

## Requeridos por dependencia (NO están en este flist; los trae su feature)

| path | dueña | nota |
|---|---|---|
| `wire/actor_providers.go` (A) | feature-007 | `ActorSet`, `ProvideActor*`. Lo referencia `wire.go`/`wire_gen.go` |
| `wire/companion_providers.go` (A) | feature-012 | `ProvideCompanionClient`, `ProvideNexusClient`, `ProvideConfigNexus`. Lo referencia `wire_gen.go`/`ai_providers.go` |
| `internal/shared/lifecycle/*` (A) | feature-002 | `RunArchiveCleanup`, `RegisterMetrics`, tipos del reporte. Lo consume `archive-cleanup` y `main.go` |
| `internal/actor/*` | feature-007 | tipo `actor.Handler` |
| `internal/ai/*`, `internal/axis/*` | feature-012 | adapter Companion |
| `internal/*/csv_exporter.go` (NewCSVExporter) | feature-013 | implementación CSV por módulo |
| `cmd/config/*` (Companion, Reporting, Security.AITenantScope) | feature-005 | secciones de config nuevas |

## Dudosos

| path | duda | cómo resolver |
|---|---|---|
| `wire/admin_providers.go` | ¿el split Repository/UseCases del admin es de 018 o de un refactor propio? | `git -C <repo> log --oneline 0972e565..777e5f6a -- internal/admin` y cruzar con flist be-018 |
| `wire/data_integrity_providers.go` | El reorden de args de `ProvideDataIntegrityUseCases` debe coincidir EXACTO con `dataintegrity.NewUseCases` de 018 | comparar firma con `internal/data-integrity/usecases.go` en 777e5f6a |

## NO traer todavía

- Ningún hunk de `wire/ai_providers.go`, `config_providers.go`, `data_integrity_providers.go`, `*_providers.go` de CSV ni `wire_gen.go`/`wire.go` **antes** de que 007/012/013/018 estén en develop: romperían el build por símbolos faltantes (`axis.CompanionClient`, `NewCSVExporter`, `ActorSet`, `SupplyRepositoryPort`).
