# 0002 Tenant Enforcement Via Column Default + NOT NULL (RLS Descoped)

Specification type: baseline current-state observed decision specification.

Status: Observed Implementation Decision.

No formal decision-record approval evidence was verified.

## Observed Implementation Decision: tenant_id Hardened By Column DEFAULT + NOT NULL, Not RLS

Per-tenant tables carry a `tenant_id uuid` column hardened by a `DEFAULT` (resolved dynamically to the `default` tenant) plus `SET NOT NULL`, applied to 11 tables: `customers`, `campaigns`, `projects`, `managers`, `investors`, `providers`, `crops`, `types`, `lease_types`, `business_parameters`, `categories`. The migration backfills `NULL` rows to the `default` tenant before the constraint, and resolves the `default` tenant uuid at runtime (portable across environments, aborting if it does not exist).

The `DEFAULT` is the stamping mechanism: the GORM models have no `TenantID` field, so GORM omits the column on INSERT and the column default supplies a non-null value without touching application code. Per-request stamping of the real tenant (Model 2, ≥2 tenants) is noted as a separate step; under Model 1 every row is `default`.

Evidence:
- `migrations_v4/000245_tenant_id_default_not_null.up.sql`
- `migrations_v4/000243_actor_tenant_not_null.up.sql` (same pattern for the actor tables)

## Observed Implementation Decision: Read Scoping Is Application-Level And Flag-Gated

Filtering by tenant is done in the application layer, gated by `TenantEnforcementEnabled()` (env `TENANT_ENFORCEMENT`, `sync.Once`, default off) and the active tenant from `OrgIDFromContext`. With the flag off the behavior is the prior one (no tenant filter) and the `tenant_id` column is not referenced in queries. The registry read path shows the pattern: it only appends `AND <alias>.tenant_id = @tenant` when both `OrgIDFromContext` returns a tenant and `TenantEnforcementEnabled()` is true.

Evidence:
- `internal/shared/models/base.go` (`OrgIDFromContext`, `TenantEnforcementEnabled`)
- `internal/registry/repository.go` (scoped/tenantSQL gating in `SearchRegistry`)

## Observed Implementation Decision: Row-Level Security Was Descoped

No `ENABLE ROW LEVEL SECURITY` / `CREATE POLICY` statement exists in `migrations_v4/`. RLS appears only as future hardening in a migration comment that lists "NOT NULL / FK VALIDATE / unicidad / RLS" as belonging to a later phase. The implemented enforcement is column DEFAULT + NOT NULL (physical integrity) plus flag-gated application scoping (read isolation), chosen over RLS. RLS remains out of scope.

Evidence:
- `migrations_v4/000232_add_tenant_id_to_workspace_roots.up.sql` (RLS named as deferred hardening, no policy statements)
- absence of `CREATE POLICY` / `ROW LEVEL SECURITY` across `migrations_v4/`

## UNKNOWN

- Whether this decision was formally approved.
- Decision owner and date.
- Whether RLS will be reconsidered when Model 2 (≥2 tenants) is activated.
- The full per-request stamping path for the real (non-`default`) tenant under Model 2.
