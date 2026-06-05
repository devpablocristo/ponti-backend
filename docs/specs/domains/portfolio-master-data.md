# Portfolio And Master Data Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Manage customers, campaigns, projects, managers, investors, providers, business parameters, FX rates, categories, and class/type catalogs.

## Boundaries

Owns:
- Customer/project/campaign reference model.
- Project manager associations.
- Investor and provider master records.
- Shared business parameters, FX rates, and classification catalogs.
- Provider lookup.

Does not own:
- Fields/lots/crops.
- Work orders.
- Stock accounting.
- Financial invoices.
- Investor allocation tables.
- Report projections.

## Owned Entities

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

## Owned APIs

- `/api/v1/customers*`
- `/api/v1/campaigns`
- `/api/v1/projects*`
- `/api/v1/managers*`
- `/api/v1/investors*`
- `/api/v1/providers`
- `/api/v1/business-parameters*`
- `/api/v1/categories*`
- `/api/v1/types*`

## Dependencies On Other Domains

- Platform, Identity, And Admin for auth/context.

## Inbound Dependencies

- Land And Crops depends on `projects`.
- Field Operations depends on `projects`, `investors`, `categories`, and `types`.
- Inventory And Stock depends on `projects`, `providers`, `investors`, `categories`, and `types`.
- Finance And Investor Accounting depends on `projects` and `investors`.
- Reporting And Data Integrity depends on `customers`, `projects`, and `campaigns`.

## Outbound Dependencies

- PostgreSQL.
- Trigram words suggester for project search.

## Aggregate Roots

- `Customer`
- `Campaign`
- `Project`
- `Manager`
- `Investor`
- `Provider`
- `BusinessParameter`
- `Category`
- `ClassType`

## Critical Business Rules

- Project name must be unique within campaign.
- Customer, campaign, manager, investor, provider, crop, lease type, and type names have DB-level unique constraints where verified.
- Project archive/restore cascades behavior to related project records in repository logic; entity ownership remains defined by `docs/specs/data/entity-ownership.md`.
- Campaign API exposes list only; create/get usecases exist but public routes are not verified.
- FX rates have a verified unique pair/date constraint, but no dedicated API ownership was verified.

## Tenant Isolation Requirements

- Business isolation is evidenced through customer/project/campaign workspace filters and FKs.
- Physical `tenant_id` is not verified on portfolio tables.

## Security Requirements

- Baseline auth applies.
- Reads require `api.read`.
- Mutations require `api.write`.

## Evidence

- `internal/customer/handler.go`
- `internal/campaign/handler.go`
- `internal/project/handler.go`
- `internal/project/usecases.go`
- `internal/project/repository.go`
- `internal/manager/handler.go`
- `internal/investor/handler.go`
- `internal/provider/handler.go`
- `internal/business-parameters/handler.go`
- `internal/category/handler.go`
- `internal/class-type/handler.go`
- `migrations_v4/000020_projects_tables.up.sql`
- `migrations_v4/000070_investors_commercialization_tables.up.sql`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`
