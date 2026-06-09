# Field Operations Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Manage labors, work orders, archived work order listing, work order items, investor splits, work order drafts, digital/batch draft workflows, draft PDF data, publication, and operational exports.

## Boundaries

Owns:
- Labor definitions per project.
- Work order lifecycle, including archive/restore and archived listing.
- Work order item usage.
- Work order investor split and payment status.
- Work order draft lifecycle and publication.

Does not own:
- Supply catalog.
- Stock ledger.
- Invoice records.
- Reporting And Data Integrity views.

## Owned Entities

- `labor_types`
- `labor_categories`
- `labors`
- `workorders`
- `workorder_items`
- `workorder_investor_splits`
- `work_order_drafts`
- `work_order_draft_items`
- `work_order_draft_investor_splits`

## Owned APIs

- `/api/v1/projects/:project_id/labors*`
- `/api/v1/labors*`
- `/api/v1/work-orders*`
- `/api/v1/work-order-drafts*`

## Dependencies On Other Domains

- Portfolio And Master Data: projects, investors, categories.
- Land And Crops: fields, lots, crops.
- Inventory And Stock: supplies.
- Platform, Identity, And Admin auth.

## Inbound Dependencies

- Finance And Investor Accounting invoices depend on work orders and work order investor membership.
- Reporting And Data Integrity read work order/labor data.
- Inventory And Stock usage counts depend on work order items.

## Outbound Dependencies

- PostgreSQL.
- XLSX export.

## Aggregate Roots

- `Labor`
- `WorkOrder`
- `WorkOrderDraft`

## Critical Business Rules

- Work order date cannot be future.
- Work order item `supply_id`, `total_used`, and `final_dose` must be positive.
- Duplicate supply items in one work order are rejected.
- Investor splits require positive investor IDs and percentages.
- Investor split percentages must sum to 100 with a small decimal tolerance.
- Investor payment status must be `Pendiente` or `Pagada`.
- Harvest effective area cannot exceed lot hectares.
- Published drafts cannot be updated or deleted.
- Draft publication fails if any referenced supply is pending.
- Digital draft numbering uses `D-n` and split suffix forms.
- Digital batch drafts with more than one lot persist one physical draft per lot
  (`D-n.1`, `D-n.2`, ...), but the batch input item `total_used` is the total
  consumption entered for the batch, not per-lot consumption.
- For digital multi-lot drafts, every lot must carry the same supply set. Core
  computes `final_dose = total_used / total_effective_area`, stores each lot as
  `total_used_lot = total_used * lot_effective_area / total_effective_area`,
  and adjusts the last lot for decimal residue so the persisted sum equals the
  original total.
- Digital group edits use the same proportional distribution. Group detail
  responses aggregate items across all physical draft lots for editing the
  batch, but published/listed work orders remain one physical row per lot.
- Work order list, filter-row, and export responses must not invent a `D-n`
  multi-lot work order row. They expose the physical split rows (`D-n.1`,
  `D-n.2`, ...), with one row per physical order/draft. Component rows from
  the report view, such as supply and labor cost rows for the same order ID, are
  aggregated into that single physical row.
- Work order duplicate endpoint is currently `Stubbed`.
- Archived work order listing reads soft-deleted work orders where `deleted_at IS NOT NULL` from the base `workorders` table using GORM `Unscoped()`.
- Archived work order listing reuses the existing work order list response mapping through `dto.FromDomainList(pageInfo, list)`.
- Project labor catalog lookup (`GET /api/v1/projects/:project_id/labors`) is
  the canonical source for labor selectors and master-data labor screens in Web
  and Mobile. It depends on migration
  `migrations_v4/000232_labor_pending_changes.up.sql` so `labors.is_pending`
  exists and pending labors can have nullable `category_id`.
- Grouped labor lookup (`GET /api/v1/labors/group/:project_id`) is a
  work-order/reporting view, not a substitute for the project labor catalog.

## Tenant Isolation Requirements

- List/filter/export workflows use workspace filters.
- Archived work order listing is verified as global and does not use workspace filters.
- FKs enforce project/field/lot/crop/labor/supply relationships.
- Physical `tenant_id` is not verified on operations tables.

## Security Requirements

- Baseline auth applies.
- Mutations require `api.write`.
- Reads/exports require `api.read`.

## Evidence

- `internal/labor/handler.go`
- `internal/labor/usecases.go`
- `internal/labor/repository_test.go`
- `internal/work-order/handler.go`
- `internal/work-order/usecases.go`
- `internal/work-order-draft/handler.go`
- `internal/work-order-draft/usecases.go`
- `migrations_v4/000050_workorders_labors_tables.up.sql`
- `migrations_v4/000190_workorder_investor_splits.up.sql`
- `migrations_v4/000202_workorder_split_payment_status.up.sql`
- `migrations_v4/000205_work_order_drafts.up.sql`
- `migrations_v4/000230_workorders_is_digital_origin.up.sql`
- `migrations_v4/000232_labor_pending_changes.up.sql`
- `migrations_v4/000233_fix_multilot_workorder_consumption.up.sql`

## Validation Evidence 2026-06-08

- `GET /api/v1/projects/30/labors` before `000232`: `500 failed to list labor`.
- Active DB after `000232` and `000233`: `schema_migrations=233`,
  `dirty=false`, `labors.category_id` nullable, `labors.is_pending` present.
- `GET /api/v1/projects/30/labors` after `000232`: `200`, 19 rows,
  `{ data, page_info }`.
- Digital multi-lot unit coverage:
  `go test ./internal/work-order-draft/...` verifies 50/50 and uneven-area
  distribution, group update distribution, decimal residue, and aggregated group
  detail items.
- Work order list coverage: `go test ./internal/work-order/...` verifies
  `D-n.1`/`D-n.2` rows remain physical split rows, each physical order appears
  once, and their summed consumption is `200`.
- Labor tests: `go test ./internal/labor/...`.
- Full Core regression: `go test ./...`.
- Runtime list check against active DB verified `D-1905555.1`, `D-1905555.2`,
  and `D-1905555.3` appear once each after component-row aggregation. The
  stored draft items have `effective_area` `201`, `35`, and `250`, `final_dose`
  `10`, and corrected `total_used` `2010`, `350`, and `2500`; the summed
  consumption is `4860`.
- The active report view still emits two component rows for each `D-1905555.x`
  physical draft row, so the API aggregation contract is required even after the
  consumption values are repaired.
