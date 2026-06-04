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

Most agricultural, operational, inventory, and finance tables do not have verified physical `tenant_id` columns.

Therefore:

- tenant isolation for these tables is not verified as physical row-level tenant isolation;
- current evidence supports workspace/project-based filtering and FK scoping;
- tenant-to-workspace ownership enforcement is UNKNOWN.

## Future Spec Requirement

Future specs must not claim full multi-tenant data isolation until current code/migrations verify it.
