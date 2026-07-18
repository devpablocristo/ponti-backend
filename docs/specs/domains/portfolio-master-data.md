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
- `/api/v1/campaigns*`
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
- Platform, Identity, And Admin reads `crops`, `types`, `lease_types`, and `campaigns` into the aggregated Registry read-model (`GET /registry`, UNION with actors). The Registry only reads these catalogs; their lifecycle and dedup remain owned here.

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
- Campaign API exposes the full lifecycle: list/create, `GET /campaigns/archived`, and per-id get/update/delete plus archive/restore (`internal/campaign/handler.go`, `internal/campaign/handler_crudar.go`).
- FX rates have a verified unique pair/date constraint, but no dedicated API ownership was verified.

### Catalog Lifecycle And Name Dedup

- Lifecycle catalogs `crops`, `types` (class-type), `lease_types` (lease-type), and `campaigns` support status filtering. List accepts `?status=active|archived|all` via `sharedrepo.ScopeByStatus` (`archived` → `deleted_at IS NOT NULL`; `all` → unscoped; default → active only). For campaigns, `?status=archived` routes to `GET /campaigns/archived`.
- Archive/restore already existed in core; the BFF `/catalog/*` mounts these handlers with `{archive:true}`.
- Name dedup by normalized name (migration 244): triggers `normalize_name`/`prevent_duplicate_name` on `crops`/`types`/`lease_types` raise a `unique_violation` on create AND on rename/reactivation; `categories` uses `prevent_duplicate_category_name`, deduping within `(tenant, type_id, normalize_name)`. `business_parameters` is excluded (dedup by `key`). Repositories map the unique-violation to HTTP 409 in both create and Update (and on restore).
- `app.bypass_name_dedup='on'` is the escape hatch; existing duplicates are not touched (only new inserts/renames/reactivations are blocked).

## Tenant Isolation Requirements

- Business isolation is evidenced through customer/project/campaign workspace filters and FKs.
- Physical `tenant_id` is now `NOT NULL` with a `default` tenant DEFAULT on `customers`, `campaigns`, `projects`, `managers`, `investors`, `providers`, `crops`, `types`, `lease_types`, `business_parameters`, and `categories` (migration 245). Stamping/scoping is flag-gated by `OrgIDFromContext`/`TenantEnforcementEnabled()`; RLS is out of scope.

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
- `internal/class-type/repository.go`
- `internal/crop/repository.go`
- `internal/lease-type/repository.go`
- `internal/shared/repository/status.go`
- `migrations_v4/000020_projects_tables.up.sql`
- `migrations_v4/000070_investors_commercialization_tables.up.sql`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`
- `migrations_v4/000244_catalog_name_dedup.up.sql`
- `migrations_v4/000245_tenant_id_default_not_null.up.sql`
