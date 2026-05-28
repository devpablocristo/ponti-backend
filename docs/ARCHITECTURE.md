## TLDR

Hexagonal pragmática + GORM (CRUD) + SQL directo (reportes/agregaciones) + observabilidad end-to-end (slog JSON, Prometheus, OTel).

**Para guía rápida orientada a Claude/onboarding**: ver [`/CLAUDE.md`](../CLAUDE.md) en root.

## Arquitectura por módulo

Layout canónico en `internal/<module>/`:

```
<module>/
├── handler.go             # HTTP routing, recibe UseCasesPort
├── handler/dto/           # Request/response (json: tags aquí)
├── usecases.go            # Orquestación + transacciones
├── usecases/domain/       # Entities sin tags; métodos + invariantes
├── repository.go          # Implementa RepositoryPort
└── repository/models/     # GORM models (column: tags); mappers ToDomain()
```

### Reglas duras

1. **Domain puro**: `usecases/domain/*.go` NO importa `gorm`, `gin`, `json`. Solo `time`, otros domain types, `domainerr`.
2. **Handler delgado**: recibe `UseCasesPort`, NO `*gorm.DB`. Solo HTTP↔Usecase marshaling.
3. **Transacciones en usecase**: `tx.Run(ctx, fn)` o equivalente; repos reciben `tx`.
4. **Errores tipados**: domain retorna `domainerr.Conflict/Validation/...`; handler mapea a HTTP via mapper centralizado en `internal/shared/handlers/errors.go`.

### Excepciones documentadas

- `internal/admin/`, `internal/businessinsights/`: pueden tener layout más plano (RBAC/proxy, no entidad de negocio típica). Admin SÍ aplica el patrón hex completo post-refactor (sept 2026). Businessinsights tiene Repository+Service exportados pero sin subcarpetas dto/.
- `internal/ai/`, `internal/reviewproxy/`: proxies hacia servicios externos. No tienen repo propio.
- `internal/data-integrity/`: cross-module reader (no tiene tabla propia).

## GORM vs SQL directo

- Usar **GORM** para operaciones CRUD simples.
- Usar **SQL directo** para consultas complejas o críticas.

## Alcance
Este documento define cuándo usar GORM y cuándo usar SQL directo en este proyecto.

## Regla principal
- Usar **GORM** para operaciones CRUD simples.
- Usar **SQL directo** para consultas complejas o críticas.

## Usar GORM cuando
- CRUD básico por ID.
- Filtros simples sin agregaciones complejas.
- Transacciones simples de escritura.
- Relaciones básicas con `Preload` y sin explosión de joins.

## Usar SQL directo cuando
- Reportes, métricas, dashboards o vistas.
- Agregaciones con `GROUP BY`, `CTE`, subqueries complejas.
- Consultas que requieren índices específicos o tuning fino.
- Performance crítica o comportamiento exacto de SQL.

## Reglas operativas
- No mezclar GORM y SQL directo en la misma función sin necesidad.
- Si una query crítica usa `Raw()`, agregar comentario breve explicando por qué.
- Evitar `Preload` en cascada si no es imprescindible.

## Documentación mínima por módulo
Cada módulo debe indicar si:
- CRUD principal usa GORM.
- Reportes o métricas usan SQL directo.

## Ejemplo de decisión
- `internal/report/*`: SQL directo.
- `internal/customer/*`: GORM CRUD.

## Observabilidad

- **Logs**: `slog` JSON a stdout via `platform/observability/go.NewJSONLogger`. Enriquecido con `request_id`, `trace_id`, `span_id`, `user_id`, `tenant_id`, `role`.
- **Métricas**: Prometheus en `GET /observability/metrics`. RED metrics HTTP por route pattern Gin + Go runtime + counter `crudar_rejected_archived_ref_total{table}`.
- **Tracing**: OTel con exporter configurable (`OTEL_EXPORTER=otlp|stdout|none`). Default `none` (no overhead).
- Ver [docs/OBSERVABILITY.md](OBSERVABILITY.md) para setup, env vars, decisiones.

## Lifecycle (CRUDAR)

Patrón "archived = no existe" + integridad jerárquica enforced:

- `lifecycle.RequireActive(tx, table, label, id)` valida que una FK no apunte a row archivado.
- `lifecycle.RunCascadeArchive/Restore` propaga lifecycle a children + pivots con `Cause` tracking.
- DB triggers (`migrations_v4/000233`) como red de seguridad última.
- Counter Prometheus `crudar_rejected_archived_ref_total{table}` mide rechazos (señal de gap UX).
- Ver [docs/crudar-lifecycle.md](crudar-lifecycle.md) y [docs/archive-restore-policy.md](archive-restore-policy.md).

## Errores

Catálogo + reglas en [docs/ERROR_CATALOG.md](ERROR_CATALOG.md). Resumen:
- Mensajes en inglés, lower-case, sin IDs/datos sensibles, máx 80 chars.
- Kinds: `Validation` (400), `Unauthorized` (401), `Forbidden` (403), `NotFound` (404), `Conflict` (409), `Unavailable` (503), `Internal` (500).
- Mapper HTTP centralizado; FE traduce a español con `translateBackendError`.

## Seguridad

- Auth: `RequireIdentityPlatformAuthz` (prod, Firebase JWT) o `RequireLocalDevAuthz` (dev/CI, header `X-USER-ID`).
- API key: `RequireAPIKeyFromEnv` en validation chain.
- CORS: `coreginmw.NewCORS` con defaults dev + env `CORS_ORIGINS`.
- Rate-limit por IP: `coreginmw.NewRateLimit` gated por `HTTP_RATE_LIMIT_PER_MINUTE`.
- Tenant scope: `authz.MaybeTenantScope` en queries (310 callers consistentes).

## AI (`InsightService` + `CopilotAgent`)
- Flujo: FE → BFF → Backend Go → Ponti AI.
- El FE no conoce claves; el Backend Go usa `X-SERVICE-KEY`.
- Ponti AI es READ-ONLY sobre dominio y solo escribe en `ai_*`.
- SQL en Ponti AI usa allowlist con `project_id` y `LIMIT` obligatorios.
- El backend expone proxy HTTP para:
  - `POST /api/v1/ai/insights/compute`
  - `GET /api/v1/ai/insights/summary`
  - `GET /api/v1/ai/insights/{entity_type}/{entity_id}`
  - `POST /api/v1/ai/insights/{insight_id}/actions`
  - `GET /api/v1/ai/copilot/insights/{insight_id}/explain`
  - `GET /api/v1/ai/copilot/insights/{insight_id}/why`
  - `GET /api/v1/ai/copilot/insights/{insight_id}/next-steps`

