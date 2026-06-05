# 0000 Observed Architectural Decisions

Specification type: baseline current-state observed decision specification.

Status: Observed Implementation Decision collection.

No formal decision-record approval evidence was verified.

## Observed Implementation Decision: Go HTTP API With Gin

The backend is implemented as a Go HTTP API using Gin/platform HTTP adapters.

Evidence:
- `go.mod`
- `cmd/api/main.go`
- `cmd/api/http_server.go`
- `internal/platform/http/servers/gin/*`

## Observed Implementation Decision: REST API Under `/api/v1`

The current API surface is registered under `/api/v1` with Gin route handlers.

Evidence:
- `cmd/api/http_server.go`
- `internal/*/handler.go`

## Observed Implementation Decision: PostgreSQL With SQL Migrations

The application uses PostgreSQL/GORM adapters and SQL migrations under `migrations_v4/`.

Evidence:
- `internal/platform/persistence/gorm/*`
- `cmd/migrate/*`
- `migrations_v4/*`

## Observed Implementation Decision: Reporting Through SQL Views/Functions

Reporting and dashboards rely on `v4_core`, `v4_ssot`, `v4_calc`, and `v4_report` schemas.

Evidence:
- `migrations_v4/000090_v4_schemas.up.sql`
- `migrations_v4/000100_v4_core_functions.up.sql`
- `migrations_v4/000110_v4_ssot_functions.up.sql`
- `migrations_v4/000120_v4_calc_views.up.sql`
- `migrations_v4/000130_v4_report_views.up.sql`

## Observed Implementation Decision: Auth Middleware Enforces API Permissions

The identity middleware maps HTTP methods to `api.read` or `api.write` and resolves role/tenant context.

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `migrations_v4/000180_authn_authz_mvp.up.sql`

## Observed Implementation Decision: Workspace-Based Business Filtering

Many reporting/list flows use `customer_id`, `project_id`, `campaign_id`, and optional `field_id` workspace filters.

Evidence:
- `internal/shared/handlers/workspace_filters.go`
- `internal/shared/filters/workspace.go`
- `internal/dashboard/handler.go`
- `internal/report/handler.go`
- `internal/work-order/usecases.go`

## Observed Implementation Decision: External AI Is Proxied, Not Implemented Locally

AI chat and conversation behavior is proxied to an external Ponti AI service, with fallback behavior when not configured.

Evidence:
- `internal/ai/client.go`
- `internal/ai/usecases/usecases.go`
- `internal/ai/handler.go`

## Observed Implementation Decision: Negative Stock Insights Are Governance-Gated

Negative real stock can create business insight candidates only after Review/Nexus allows the request.

Evidence:
- `internal/stock/usecases.go`
- `internal/businessinsights/service.go`
- `internal/reviewproxy/client.go`

## Observed Implementation Decision: GitHub Actions Drive CI/CD

Build, deploy, rollback, migration, and DB reset automation are represented as GitHub workflow files.

Evidence:
- `.github/workflows/*`

## UNKNOWN

- Whether any of these observed implementation decisions were formally approved.
- Historical rationale for the choices.
- Decision owners.
- Superseded decisions.
