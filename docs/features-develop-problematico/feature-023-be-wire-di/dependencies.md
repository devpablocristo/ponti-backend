# dependencies.md — feature-023 · be-wire-di

Esta feature es **costura** (Wire DI + bootstrap). Por naturaleza depende de casi toda la oleada y no bloquea a casi nadie funcionalmente, pero es el último eslabón de compilación.

## Depende de

### Fuertes (sin esto NO compila)
| feature | qué aporta | símbolo/archivo concreto |
|---|---|---|
| 001 be-platform-tenancy-refactor | paquetes `platform/*` | `github.com/devpablocristo/platform/observability/go`, `platform/http/gin/go` (CORS, RateLimit) |
| 002 be-crudar-lifecycle-framework | runtime de cleanup/metrics | `internal/shared/lifecycle.RunArchiveCleanup`, `lifecycle.RegisterMetrics`, `ArchiveCleanupReport`/`Options` |
| 005 be-config-modularization | secciones de config | `config.Companion`, `config.Reporting.ReadMode`, `config.Security.AITenantScope`, `cfg.Auth.RequireTenantHeader`, `cfg.Service.Env` |
| 007 actor-system | módulo actor + providers | `internal/actor.Handler`, `wire/actor_providers.go` (`ActorSet`, `ProvideActor*`) |
| 012 ai-companion-integration | cliente companion + providers | `internal/axis.CompanionClient`, `ai.NewCompanionAdapter`, `wire/companion_providers.go` (`ProvideCompanionClient`, `ProvideNexusClient`, `ProvideConfigNexus`) |
| 013 be-csv-export | exporters CSV | `lot/supply/stock/work-order/labor.NewCSVExporter` (reemplazan XLSX engine) |
| 018 data-integrity-admin | ports + firma use-cases | `dataintegrity.SupplyRepositoryPort`, `ProjectRepositoryPort`, firma de `NewUseCases(dashboard, workorder, report, supply, project, lot)`; split admin Repository/UseCases |

### Débiles (cablea pero podría stubbearse)
| feature | nota |
|---|---|
| 008 identity-tenant-context | `middleware_providers.go` agrega `Environment` + `RequireTenantHeader` al auth config; si 008 no está, esos campos no existen en config |
| 009 crudar-archive-surface | `archive-cleanup` recorre la superficie de borrado/archive; sin 009 el cleanup no tiene qué limpiar (pero compila) |

### Inciertas
| feature | duda |
|---|---|
| 018 vs refactor admin | el split `ProvideAdminRepository`/`ProvideAdminUseCases` podría ser parte de admin/018 o de un refactor independiente — verificar con `git log 0972e565..777e5f6a -- internal/admin` |

## Bloquea a
- A nadie funcionalmente. Pero **debe mergearse al final** de la oleada porque `wire_gen.go` referencia a todos los módulos: si 023 entra antes que 007/012/013/018, el build de develop queda roto. Inversamente, 007/012/013/018 deben traer sus propios hunks de `wire/*` para no dejar el grafo a medias.

## Archivos / tipos / config / APIs compartidos
- **Archivos costura (editados por múltiples features)**: `wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go`, `cmd/api/http_server.go`, `wire/config_providers.go`, `wire/middleware_providers.go`, `wire/ai_providers.go`, `wire/data_integrity_providers.go`.
- **Tipos compartidos**: `wire.Dependencies` (campo `ActorHandler`), `MiddlewaresEnginePort` (sin `GetProtected`).
- **Config compartida**: `cmd/config` secciones Companion/Reporting/Security/Auth (005).
- **APIs**: `GET /observability/metrics` nuevo; registro de rutas actor.

## Cross-repo
- Ninguna. Solo-BE. En FE: "sin cambios".

## Recomendación de orden
```
001  → 005  → 002 / 008 / 009  → 007 / 012 / 013 / 018  → 023 (cierre de costura)
```
Cada una de 007/012/013/018 trae SUS hunks de `wire/*` con `restore -p`; 023 aporta el bootstrap (observability, slog, archive-cleanup) y, si hace falta, regenera `wire_gen.go`.
