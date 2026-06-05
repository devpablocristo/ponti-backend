# Schema Inventory Baseline Specification

Specification type: baseline current-state data specification.

## Verified Schemas

- `public`
- `v4_core`
- `v4_ssot`
- `v4_calc`
- `v4_report`

Evidence:
- `migrations_v4/000090_v4_schemas.up.sql`

## Core Public Tables

Platform, Identity, And Admin:
- `users`
- `auth_tenants`
- `auth_roles`
- `auth_permissions`
- `auth_role_permissions`
- `auth_memberships`

Portfolio And Master Data:
- `customers`
- `campaigns`
- `projects`
- `managers`
- `project_managers`
- `investors`
- `providers`
- `business_parameters`
- `fx_rates`
- `categories`
- `types`

Land And Crops:
- `lease_types`
- `fields`
- `lots`
- `lot_dates`
- `crops`

Field Operations:
- `labor_types`
- `labor_categories`
- `labors`
- `workorders`
- `workorder_items`
- `workorder_investor_splits`
- `work_order_drafts`
- `work_order_draft_items`
- `work_order_draft_investor_splits`

Inventory And Stock:
- `supplies`
- `stocks`
- `supply_movements`

Finance And Investor Accounting:
- `project_investors`
- `field_investors`
- `admin_cost_investors`
- `crop_commercializations`
- `project_dollar_values`
- `invoices`

AI And Business Insights:
- `business_insight_candidates`
- `business_insight_reads`

## Evidence

- `migrations_v4/000010_core_tables.up.sql`
- `migrations_v4/000020_projects_tables.up.sql`
- `migrations_v4/000030_fields_lots_tables.up.sql`
- `migrations_v4/000040_crops_tables.up.sql`
- `migrations_v4/000050_workorders_labors_tables.up.sql`
- `migrations_v4/000060_supplies_inventory_tables.up.sql`
- `migrations_v4/000070_investors_commercialization_tables.up.sql`
- `migrations_v4/000180_authn_authz_mvp.up.sql`
- `migrations_v4/000190_workorder_investor_splits.up.sql`
- `migrations_v4/000205_work_order_drafts.up.sql`
- `migrations_v4/000209_business_insight_candidates.up.sql`
- `migrations_v4/000210_business_insight_reads.up.sql`

## UNKNOWN

- Any tables created outside `migrations_v4/`.
- Any production-only schemas not represented in migrations.
