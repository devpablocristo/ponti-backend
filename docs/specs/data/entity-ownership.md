# Entity Ownership Baseline Specification

Specification type: baseline current-state data ownership specification.

| Domain | Owned Entities |
|---|---|
| Platform, Identity, And Admin | `users`, `auth_tenants`, `auth_roles`, `auth_permissions`, `auth_role_permissions`, `auth_memberships` |
| Runtime, Migration, And Delivery | No business entities |
| Portfolio And Master Data | `customers`, `campaigns`, `projects`, `managers`, `project_managers`, `investors`, `providers`, `business_parameters`, `fx_rates`, `categories`, `types` |
| Land And Crops | `lease_types`, `fields`, `lots`, `lot_dates`, `crops` |
| Field Operations | `labor_types`, `labor_categories`, `labors`, `workorders`, `workorder_items`, `workorder_investor_splits`, `work_order_drafts`, `work_order_draft_items`, `work_order_draft_investor_splits` |
| Inventory And Stock | `supplies`, `stocks`, `supply_movements` |
| Finance And Investor Accounting | `project_investors`, `field_investors`, `admin_cost_investors`, `crop_commercializations`, `project_dollar_values`, `invoices` |
| Reporting And Data Integrity | `v4_core`, `v4_ssot`, `v4_calc`, `v4_report` projections |
| AI And Business Insights | `business_insight_candidates`, `business_insight_reads` |

## Shared / Cross-Domain Entities

- `investors` are owned by Portfolio And Master Data but used by Finance And Investor Accounting, Field Operations, and Inventory And Stock.
- `providers` are owned by Portfolio And Master Data; Inventory And Stock may reference them and may create them through verified movement/import flows.
- `projects` are owned by Portfolio And Master Data and anchor most business data.
- `field_investors`, `project_investors`, and `admin_cost_investors` are owned by Finance And Investor Accounting and referenced by Land And Crops, Portfolio And Master Data, and Reporting And Data Integrity views where applicable.
- `supplies` are owned by Inventory And Stock and referenced by Field Operations.
- `workorders` are owned by Field Operations and referenced by Finance And Investor Accounting and Reporting And Data Integrity.

## Evidence

- `internal/*/repository*.go`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`
