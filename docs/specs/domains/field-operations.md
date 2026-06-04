# Field Operations Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Manage labors, work orders, work order items, investor splits, work order drafts, digital/batch draft workflows, draft PDF data, publication, and operational exports.

## Boundaries

Owns:
- Labor definitions per project.
- Work order lifecycle.
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
- Work order duplicate endpoint is currently `Stubbed`.

## Tenant Isolation Requirements

- List/filter/export workflows use workspace filters.
- FKs enforce project/field/lot/crop/labor/supply relationships.
- Physical `tenant_id` is not verified on operations tables.

## Security Requirements

- Baseline auth applies.
- Mutations require `api.write`.
- Reads/exports require `api.read`.

## Evidence

- `internal/labor/handler.go`
- `internal/labor/usecases.go`
- `internal/work-order/handler.go`
- `internal/work-order/usecases.go`
- `internal/work-order-draft/handler.go`
- `internal/work-order-draft/usecases.go`
- `migrations_v4/000050_workorders_labors_tables.up.sql`
- `migrations_v4/000190_workorder_investor_splits.up.sql`
- `migrations_v4/000202_workorder_split_payment_status.up.sql`
- `migrations_v4/000205_work_order_drafts.up.sql`
- `migrations_v4/000230_workorders_is_digital_origin.up.sql`
