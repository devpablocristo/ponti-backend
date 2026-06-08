# Reporting And Data Integrity Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Expose dashboard, reports, reporting views, and data integrity checks across business source data.

## Boundaries

Owns:
- Dashboard API.
- Report APIs.
- Data integrity APIs.
- Reporting read models and projections under `v4_*` schemas.

Does not own:
- Transactional source-of-truth mutations.
- Source entity lifecycle.

## Owned Entities

No transactional entities verified.

Owned/read projection schemas:
- `v4_core`
- `v4_ssot`
- `v4_calc`
- `v4_report`

## Owned APIs

- `GET /api/v1/dashboard`
- `GET /api/v1/reports/:type`
- `GET /api/v1/data-integrity/costs-check`
- `GET /api/v1/data-integrity/tentative-prices`

## Dependencies On Other Domains

- Portfolio And Master Data.
- Land And Crops.
- Field Operations.
- Inventory And Stock.
- Finance And Investor Accounting.
- Platform, Identity, And Admin auth.

## Inbound Dependencies

- Frontend/BFF/report consumers are expected consumers, but implementation is external and UNKNOWN.
- No transactional domain dependency on reporting was verified.

## Outbound Dependencies

- PostgreSQL views/functions.

## Aggregate Roots

No formal aggregate root verified.

Projection roots:
- `DashboardData`
- `FieldCropReport`
- `InvestorContributionReport`
- `SummaryResults`
- `IntegrityReport`

## Critical Business Rules

- Dashboard requires `customer_id`, `project_id`, and `campaign_id`; `field_id` is optional.
- Summary results requires complete workspace.
- Report type must be one of the implemented report types.
- Data integrity cost check compares values from lots, field/crop report, summary results, dashboard, work orders, investor report, and stocks.
- Tentative prices reads partial price flags.

## Tenant Isolation Requirements

- Reporting APIs use workspace filters.
- Physical tenant-scoped reporting views are not verified.
- Tenant safety depends on correct workspace-to-tenant enforcement, which is UNKNOWN.

## Security Requirements

- Baseline auth applies.
- APIs are read-only and require `api.read`.

## Evidence

- `internal/dashboard/handler.go`
- `internal/dashboard/repository.go`
- `internal/report/handler.go`
- `internal/report/usecases.go`
- `internal/report/usecases/validators.go`
- `internal/data-integrity/handler.go`
- `internal/data-integrity/usecases.go`
- `migrations_v4/000090_v4_schemas.up.sql`
- `migrations_v4/000100_v4_core_functions.up.sql`
- `migrations_v4/000110_v4_ssot_functions.up.sql`
- `migrations_v4/000120_v4_calc_views.up.sql`
- `migrations_v4/000130_v4_report_views.up.sql`
