# Inventory And Stock Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Manage supply catalog, pending supplies, supply movements, stock movements, internal transfers, returns, real stock counts, stock periods, and stock export.

## Boundaries

Owns:
- Supplies.
- Pending supply lifecycle.
- Supply movement and stock movement API surfaces.
- Stock period close and carry-forward.
- Real stock count mutation.

Does not own:
- Work order creation.
- Financial invoices.
- Business insight candidate storage, though it can trigger insight notifications.

## Owned Entities

- `supplies`
- `stocks`
- `supply_movements`

## Owned APIs

- `/api/v1/supplies*`
- `/api/v1/projects/:project_id/supply-movements*`
- `/api/v1/projects/:project_id/stock-movements*`
- `/api/v1/projects/:project_id/stocks*`

## Dependencies On Other Domains

- Portfolio And Master Data: projects, investors, providers, categories, types.
- Field Operations: work order item usage counts.
- AI And Business Insights: optional negative stock notifier.
- Platform, Identity, And Admin auth.

## Inbound Dependencies

- Field Operations uses supplies in work orders and drafts.
- Reporting And Data Integrity read supplies/stocks/movements.
- AI And Business Insights receives stock-negative notification attempts.

## Outbound Dependencies

- PostgreSQL.
- XLSX import/export.
- Optional Business Insights notifier.

## Aggregate Roots

- `Supply`
- `SupplyMovement`
- `Stock`

## Critical Business Rules

- Supply create/update requires project, name, unit, category, and type.
- Pending supply creation normalizes name and reuses existing project/name matches.
- Bulk supply creation requires all supplies share a project and request names are unique.
- Completing pending supply requires the current supply to still be pending.
- Movement supply must belong to movement project.
- Investor must exist.
- Provider must exist or a provider name must be supplied; provider master records remain owned by Portfolio And Master Data even when created through movement/import flows.
- Duplicate reference/supply combinations are rejected.
- Manual stock count updates real stock and does not create a movement record.
- Return movement requires existing stock and cannot exceed available stock.
- Internal movement requires destination project, destination different from source, and creates out/in records.
- Strict mode is transactional.
- Stock close carries forward active and eligible closed stocks.

## Tenant Isolation Requirements

- Business data is project-scoped.
- Negative stock insight emission uses auth `OrgID` when available.
- Physical `tenant_id` is not verified on supply/stock tables.

## Security Requirements

- Baseline auth applies.
- Mutations/import/close/real-stock require `api.write`.
- Reads/exports require `api.read`.

## Evidence

- `internal/supply/handler.go`
- `internal/supply/usecases.go`
- `internal/supply/usecases_movement.go`
- `internal/supply/usecases_movement_import.go`
- `internal/stock/handler.go`
- `internal/stock/usecases.go`
- `migrations_v4/000060_supplies_inventory_tables.up.sql`
- `migrations_v4/000197_supply_partial_price_flag.up.sql`
- `migrations_v4/000199_supply_return_movement_type.up.sql`
- `migrations_v4/000203_stock_real_count_flag.up.sql`
- `migrations_v4/000211_supply_pending_flag.up.sql`
