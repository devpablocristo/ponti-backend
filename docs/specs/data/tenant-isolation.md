# Tenant Isolation Data Baseline Specification

Specification type: baseline current-state data isolation specification.

## Verified Tenant-Scoped Data

`business_insight_candidates` has `tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id)` and unique candidate fingerprint by tenant.

Evidence:
- `migrations_v4/000209_business_insight_candidates.up.sql`
- `internal/businessinsights/repository.go`

## Verified Auth Tenant Model

Auth tables define tenants, roles, permissions, role permissions, and memberships.

Evidence:
- `migrations_v4/000180_authn_authz_mvp.up.sql`
- `migrations_v4/000201_auth_uuid_migration.up.sql`

## Verified Physical `tenant_id` Columns (Pilar 1)

The following 11 master/catalog tables have `tenant_id` with `DEFAULT 'default'` and `NOT NULL`: `customers`, `campaigns`, `projects`, `managers`, `investors`, `providers`, `crops`, `types`, `lease_types`, `business_parameters`, `categories`. The default tenant is resolved dynamically by name.

`actors` and `actor_keys` (Identity Gate / registry) also have `tenant_id NOT NULL`, with per-tenant dedup enforced by the `uq_actor_keys_active (tenant_id, key_type, key_value) WHERE active` partial unique index. See `data-model-actors.md`.

Physical filtering by `tenant_id` is flag-gated: it is applied only when `TenantEnforcementEnabled()` is on, scoping by the org id from context (`OrgIDFromContext`). Under Modelo 1 every row is the `default` tenant; the column DEFAULT keeps GORM inserts (which omit the column) non-null without code changes. The registry / actors search and lookups scope by the same tenant resolution (`identity.TenantFor`), also flag-gated.

Evidence:
- `migrations_v4/000245_tenant_id_default_not_null.up.sql`
- `migrations_v4/000243_actor_tenant_not_null.up.sql`
- `internal/shared/models/base.go`
- `internal/identity/resolver.go`

## Workspace-Based Business Isolation

Most business APIs and report queries use workspace/project filters:

- `customer_id`
- `project_id`
- `campaign_id`
- optional `field_id`

Evidence:
- `internal/shared/handlers/workspace_filters.go`
- `internal/shared/filters/workspace.go`
- `internal/dashboard/handler.go`
- `internal/report/handler.go`
- `internal/work-order/usecases.go`

## Critical Limitation

The 11 master/catalog tables and the actors/registry tables (above) carry physical `tenant_id`. The remaining agricultural, operational, inventory, and finance tables (e.g. `fields`, `lots`, `workorders`, `labors`, `supplies`, `stocks`, the investor-accounting tables) do not have verified physical `tenant_id` columns.

Therefore:

- tenant isolation for those remaining tables is not verified as physical row-level tenant isolation;
- current evidence supports workspace/project-based filtering and FK scoping for them;
- physical `tenant_id` filtering is flag-gated and (for the tables that have the column) only enforced when `TenantEnforcementEnabled()` is on;
- RLS (row-level security) is descoped;
- tenant-to-workspace ownership enforcement is UNKNOWN.

## Future Spec Requirement

Future specs must not claim full multi-tenant data isolation until current code/migrations verify it.
