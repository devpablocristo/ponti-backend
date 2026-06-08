# Finance And Investor Accounting Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Manage project dollar values, crop commercializations, investor allocations, and invoices per work order/investor.

## Boundaries

Owns:
- Monthly project dollar values.
- Crop commercialization price inputs and net price calculation.
- Invoice records.
- Investor allocation tables where persisted.

Does not own:
- Work order execution.
- Report view construction.
- External accounting integrations.

## Owned Entities

- `project_dollar_values`
- `crop_commercializations`
- `invoices`
- `project_investors`
- `field_investors`
- `admin_cost_investors`

## Owned APIs

- `/api/v1/projects/:project_id/dollar-values`
- `/api/v1/projects/:project_id/commercializations`
- `/api/v1/invoices*`

Dedicated APIs for allocation tables are UNKNOWN. Allocation data is verified in persistence/models and migrations, with nested project/field usage.

## Dependencies On Other Domains

- Portfolio And Master Data: projects and investors.
- Land And Crops: crops and fields.
- Field Operations: work orders and investor splits.
- Platform, Identity, And Admin auth.

## Inbound Dependencies

- Reporting And Data Integrity consume finance/investor data.
- Dashboard consumes project/admin cost and contribution views.

## Outbound Dependencies

- PostgreSQL.

## Aggregate Roots

- `ProjectDollarValue`
- `CropCommercialization`
- `Invoice`
- `InvestorAllocation`

## Critical Business Rules

- Dollar values list requires project ID.
- Dollar values bulk upsert requires all items share project and year.
- Duplicate month in a dollar values request is rejected.
- Commercialization calculates net price before persistence.
- Invoice target requires valid work order ID and investor ID.
- Invoice investor must belong to the work order.
- Invoices are unique by work order and investor after migration.

UNKNOWN:
- Complete invariant for allocation percentage totals across `project_investors`, `field_investors`, and `admin_cost_investors`.

## Tenant Isolation Requirements

- Project/workorder-scoped isolation through FKs.
- Physical `tenant_id` is not verified on finance tables.

## Security Requirements

- Baseline auth applies.
- Mutations require `api.write`.
- Reads require `api.read`.

## Evidence

- `internal/dollar/handler.go`
- `internal/dollar/usecases.go`
- `internal/commercialization/handler.go`
- `internal/commercialization/usecases.go`
- `internal/invoice/handler.go`
- `internal/invoice/usecase.go`
- `internal/project/repository.go`
- `migrations_v4/000070_investors_commercialization_tables.up.sql`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`
- `migrations_v4/000204_invoice_per_investor.up.sql`
