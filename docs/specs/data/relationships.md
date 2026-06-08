# Relationships Baseline Specification

Specification type: baseline current-state data relationship specification.

## Portfolio And Master Data Relationships

- `projects.customer_id -> customers.id`
- `projects.campaign_id -> campaigns.id`
- `project_managers.project_id -> projects.id`
- `project_managers.manager_id -> managers.id`

## Land And Crops Relationships

- `fields.project_id -> projects.id`
- `fields.lease_type_id -> lease_types.id`
- `lots.field_id -> fields.id`
- `lots.current_crop_id -> crops.id`
- `lots.previous_crop_id -> crops.id`
- `lot_dates.lot_id -> lots.id`

## Field Operations Relationships

- `labors.project_id -> projects.id`
- `labors.category_id -> categories.id`
- `workorders.project_id -> projects.id`
- `workorders.field_id -> fields.id`
- `workorders.lot_id -> lots.id`
- `workorders.crop_id -> crops.id`
- `workorders.labor_id -> labors.id`
- `workorder_items.workorder_id -> workorders.id`
- `workorder_items.supply_id -> supplies.id`
- `workorder_investor_splits.workorder_id -> workorders.id`
- `workorder_investor_splits.investor_id -> investors.id`

## Inventory And Stock Relationships

- `supplies.project_id -> projects.id` is verified by model/usecase behavior; direct FK in baseline migration was not found in the cited FK migration.
- `supplies.category_id -> categories.id`
- `supplies.type_id -> types.id`
- `stocks.project_id -> projects.id`
- `stocks.supply_id -> supplies.id`
- `stocks.investor_id -> investors.id`
- `supply_movements.supply_id -> supplies.id`
- `supply_movements.provider_id -> providers.id`
- `supply_movements.investor_id -> investors.id`

## Finance And Investor Accounting Relationships

- `project_investors.project_id -> projects.id`
- `project_investors.investor_id -> investors.id`
- `field_investors.field_id -> fields.id`
- `field_investors.investor_id -> investors.id`
- `admin_cost_investors.project_id -> projects.id`
- `admin_cost_investors.investor_id -> investors.id`
- `crop_commercializations.project_id -> projects.id`
- `crop_commercializations.crop_id -> crops.id`
- `project_dollar_values.project_id -> projects.id`
- `invoices.work_order_id -> workorders.id`
- `invoices.investor_id -> investors.id`

## AI And Business Insights Relationships

- `business_insight_candidates.tenant_id -> auth_tenants.id`
- `business_insight_reads.insight_id -> business_insight_candidates.id`

## Evidence

- `migrations_v4/000080_constraints_fks_indexes.up.sql`
- `migrations_v4/000190_workorder_investor_splits.up.sql`
- `migrations_v4/000204_invoice_per_investor.up.sql`
- `migrations_v4/000205_work_order_drafts.up.sql`
- `migrations_v4/000209_business_insight_candidates.up.sql`
- `migrations_v4/000210_business_insight_reads.up.sql`

## UNKNOWN

- Any relationships enforced only outside the repository.
- Any production-only constraints not represented in migrations.
