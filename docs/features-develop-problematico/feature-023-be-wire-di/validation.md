# validation.md — feature-023 · be-wire-di

## Checklist pre-PR (BE)
- [ ] Dependencias en develop: 001, 002, 005, 007, 008, 009, 012, 013, 018.
- [ ] `go build ./...` verde (sin `undefined:` ni imports sin usar).
- [ ] `go vet ./wire/... ./cmd/...` limpio.
- [ ] `wire_gen.go` regenerado y coherente: `go run github.com/google/wire/cmd/wire ./wire` no reporta diff inesperado ni "no provider found".
- [ ] `grep -rn "GetProtected" internal/ cmd/ wire/` → sin resultados (método eliminado de `MiddlewaresEnginePort`).
- [ ] `grep -rn "ProvideConfigAI\|ProvideAIClient\|NewExcelExporter\|XLSXEnginePort" wire/` → sin resultados (reemplazados por Companion/CSV).
- [ ] `git diff --check` (sin trailing whitespace / conflict markers).

## Tests sugeridos (BE)
- `go test ./cmd/migrate/...` (incluye `TestMigrationLock` con el nuevo `*slog.Logger`).
- `go test ./internal/shared/lifecycle/...` (validar que `RunArchiveCleanup` exista y pase — pertenece a 002, pero `archive-cleanup` lo consume).
- Test de humo manual de wire: `go run ./cmd/api` y verificar arranque sin error de `Initialize()`.

## Validación manual
1. **Arranque API**: `go run ./cmd/api` con env mínimas:
   - DB (`DB_HOST/USER/PASSWORD/NAME/PORT/SSLMODE`), `COMPANION_BASE_URL`, `COMPANION_INTERNAL_JWT_SECRET`.
   - Esperado: log JSON `"event":"http_server_starting"` con `port/version/database`.
2. **Métricas**: `curl -s localhost:8080/observability/metrics | head` → texto Prometheus (200).
3. **Healthcheck**: `curl -s localhost:8080/` → JSON con `reporting.read_mode` presente.
4. **Tracing opcional**: setear `OTEL_EXPORTER=stdout` y verificar que no aborta; con valor inválido debe loguear warn y seguir (no-op).
5. **archive-cleanup dry-run**: `go run ./cmd/archive-cleanup --dry-run --output json` → JSON con `mode/tenant/checks/actions`. Sin mutaciones.
6. **archive-cleanup flags excluyentes**: `go run ./cmd/archive-cleanup --apply --dry-run` → exit code 2 + mensaje en stderr.
7. **migrate slog**: `go run ./cmd/migrate` → logs JSON `"event":"migration_lock_acquired"`, `"migrations_completed"`.

## Casos borde
- Companion env ausentes → `Initialize()` falla con error claro (arranque abortado): comportamiento esperado, no bug.
- `--tenant-id` inválido (no-UUID) en archive-cleanup → exit 2.
- `--output` distinto de `table|json` → exit 2 (`ErrArchiveCleanupUnsupportedOutput`).
- `HTTP_RATE_LIMIT_PER_MINUTE=0` → middleware no se registra (sin rate-limit).

## Qué revisar en UI / API / DB / env
- **UI**: nada (solo-BE).
- **API**: ruta nueva `/observability/metrics`; rutas actor registradas (`deps.ActorHandler.Routes()`).
- **DB**: `archive-cleanup --apply` muta; correr siempre dry-run antes. `cmd/migrate` no cambia esquema por sí mismo.
- **Env**: `OTEL_*`, `CORS_ORIGINS`, `HTTP_RATE_LIMIT_PER_MINUTE`, `COMPANION_*`, `AI_TENANT_SCOPE` (Security.AITenantScope), `SERVICE_VERSION`, `ENVIRONMENT`.

## Qué validar en el otro repo
- Nada. Confirmar en cross-repo-map del FE que figura como "feature-023 sin cambios FE".

## Señales de incompletitud / incompatibilidad
- Build error `undefined: ActorSet` → falta 007.
- `undefined: axis.CompanionClient` / `ProvideCompanionClient` → falta 012.
- `undefined: NewCSVExporter` (lot/supply/stock/work-order/labor) → falta 013.
- `undefined: dataintegrity.SupplyRepositoryPort` o args de `NewUseCases` mal ordenados → falta/desfase con 018.
- `undefined: lifecycle.RunArchiveCleanup` → falta 002.
- `cfg.Companion`/`cfg.Reporting`/`cfg.Security.AITenantScope` undefined → falta 005.
- `wire: no provider found` al regenerar → un `*Set`/provider quedó sin traer.
