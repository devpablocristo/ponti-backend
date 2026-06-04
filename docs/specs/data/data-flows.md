# Data Flows Baseline Specification

Specification type: baseline current-state data-flow specification.

## Portfolio And Master Data To Land And Crops Flow

Projects are created under customers and campaigns. Fields reference projects. Lots reference fields and crops.

Evidence:
- `migrations_v4/000020_projects_tables.up.sql`
- `migrations_v4/000030_fields_lots_tables.up.sql`
- `migrations_v4/000040_crops_tables.up.sql`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`

## Field Operations Flow

Work orders reference project, field, lot, crop, labor, investor, and work order items. Items reference supplies. Work order investor splits can allocate work order participation and payment status.

Evidence:
- `internal/work-order/usecases.go`
- `migrations_v4/000050_workorders_labors_tables.up.sql`
- `migrations_v4/000190_workorder_investor_splits.up.sql`
- `migrations_v4/000202_workorder_split_payment_status.up.sql`

## Draft Publication Flow

Work order drafts store pending operational data. Publishing a draft creates a real work order and marks the draft as published. Publication rejects drafts with pending supplies.

Evidence:
- `internal/work-order-draft/usecases.go`
- `migrations_v4/000205_work_order_drafts.up.sql`

## Inventory And Stock Flow

Supplies are project-scoped. Supply movements create or update stock context. Internal transfers create source and destination movement records. Return movements require available stock. Real stock updates can trigger business insight evaluation.

Evidence:
- `internal/supply/usecases.go`
- `internal/supply/usecases_movement.go`
- `internal/stock/usecases.go`

## Finance And Investor Accounting Flow

Project dollar values and crop commercializations provide financial inputs. Invoices attach to work order and investor. Investor contribution calculations are projected through reporting views.

Evidence:
- `internal/dollar/usecases.go`
- `internal/commercialization/usecases.go`
- `internal/invoice/usecase.go`
- `migrations_v4/000120_v4_calc_views.up.sql`
- `migrations_v4/000130_v4_report_views.up.sql`

## Reporting And Data Integrity Flow

Reporting APIs read from `v4_*` views/functions that aggregate source tables across portfolio, land, operations, inventory, and finance.

Evidence:
- `internal/dashboard/repository.go`
- `internal/report/repository.go`
- `internal/data-integrity/usecases.go`

## AI And Business Insights Flow

Real stock update evaluates negative stock. If quantity is negative and Review/Nexus allows, a tenant-scoped business insight candidate is recorded. If stock returns non-negative, existing candidates can be resolved.

Evidence:
- `internal/stock/usecases.go`
- `internal/businessinsights/service.go`
- `internal/businessinsights/repository.go`

## UNKNOWN

- Frontend/BFF data flow.
- External AI persistence flow.
- External Review/Nexus policy evaluation internals.
