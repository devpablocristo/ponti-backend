# System Overview Baseline Specification

Specification type: baseline current-state specification.

## Purpose

The repository implements the Ponti backend API. Verified runtime evidence shows a Go HTTP API that manages agricultural project portfolio data, land/crop structure, work orders, supplies, stock, finance inputs, reporting, data integrity checks, AI proxying, and business insights.

Evidence:
- `cmd/api/main.go`
- `cmd/api/http_server.go`
- `internal/*/handler.go`
- `migrations_v4/*`

## Runtime Architecture

- Language/runtime: Go module `github.com/devpablocristo/ponti-backend`.
- HTTP framework: Gin through platform HTTP server adapters.
- Entry point: `cmd/api/main.go`.
- Dependency wiring: `wire.Initialize()`.
- Primary API base: `/api/v1`.
- Persistence: PostgreSQL via GORM/platform persistence adapters.
- Migrations: SQL files under `migrations_v4/`.

Evidence:
- `go.mod`
- `cmd/api/main.go`
- `cmd/api/http_server.go`
- `wire/`
- `internal/platform/persistence/gorm/*`
- `cmd/migrate/*`
- `migrations_v4/*`

## Verified Domain Areas

- Platform, Identity, And Admin
- Portfolio And Master Data
- Land And Crops
- Field Operations
- Inventory And Stock
- Finance And Investor Accounting
- Reporting And Data Integrity
- AI And Business Insights
- Runtime, Migration, And Delivery

Evidence:
- `internal/admin/`
- `internal/customer/`, `internal/project/`, `internal/campaign/`, `internal/manager/`, `internal/investor/`
- `internal/field/`, `internal/lot/`, `internal/crop/`, `internal/lease-type/`
- `internal/labor/`, `internal/work-order/`, `internal/work-order-draft/`
- `internal/supply/`, `internal/stock/`
- `internal/dollar/`, `internal/commercialization/`, `internal/invoice/`
- `internal/dashboard/`, `internal/report/`, `internal/data-integrity/`
- `internal/ai/`, `internal/businessinsights/`

## API Surface

The verified API surface is REST over HTTP. No current runtime GraphQL API, gRPC API, queue consumer, queue producer, or event contract was verified.

Evidence:
- `cmd/api/http_server.go`
- `internal/*/handler.go`
- `go.mod`

UNKNOWN:
- Any API surface implemented outside this repository.
- Frontend/BFF behavior.
- External AI service API internals beyond the paths proxied by this backend.

## Persistence Architecture

The repository uses PostgreSQL schemas/tables and database views/functions.

Verified schemas:
- `public`
- `v4_core`
- `v4_ssot`
- `v4_calc`
- `v4_report`

Evidence:
- `migrations_v4/000090_v4_schemas.up.sql`
- `migrations_v4/000100_v4_core_functions.up.sql`
- `migrations_v4/000110_v4_ssot_functions.up.sql`
- `migrations_v4/000120_v4_calc_views.up.sql`
- `migrations_v4/000130_v4_report_views.up.sql`

## Deployment And Runtime

Verified runtime/deployment artifacts exist for Docker, Docker Compose, GitHub Actions, migrations, GCP/Cloud Run oriented deployment, rollback, and DB reset workflows.

Evidence:
- `Dockerfile`
- `Dockerfile.dev`
- `docker-compose.yml`
- `Makefile`
- `.github/workflows/deploy-dev.yml`
- `.github/workflows/deploy-staging.yml`
- `.github/workflows/deploy-prod.yml`
- `.github/workflows/rollback-staging.yml`
- `.github/workflows/rollback-prod.yml`
- `.github/workflows/reset-dev-db-from-prod.yml`
- `.github/workflows/reset-stg-db-from-prod.yml`

UNKNOWN:
- Production topology beyond workflow/config evidence.
- Runtime monitoring dashboards.
- On-call ownership.
- Service-level objectives.
