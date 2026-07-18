# Platform, Identity, And Admin Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Provide API authentication, authorization, tenant context, user/tenant/membership administration, the central identity registry (actors), and runtime health/version surfaces.

## Boundaries

Owns:
- API validation middleware.
- API key and JWT/local auth behavior.
- Role/permission resolution (coarse roles + fine permissions).
- Admin tenant/user/membership endpoints.
- Auth tables and membership tables.
- Identity Gate: actor registry (`actors`/`actor_keys`/`actor_roles`), identity resolution/dedup, and the aggregated read-only Registry surface (`/registry`).

Does not own:
- Frontend/BFF identity behavior.
- Catalog/master-data lifecycle (owned by Portfolio And Master Data); the Registry only reads catalogs.
- Actor merge (not implemented).

## Owned Entities

- `users`
- `auth_tenants`
- `auth_roles`
- `auth_permissions`
- `auth_role_permissions`
- `auth_memberships`
- `actors`
- `actor_keys`
- `actor_roles`

### Identity Registry Tables (Pilar 3)

- `actors` (`id` bigserial, `tenant_id` uuid, `party_type` IN (`org`,`person`,`unknown`) default `unknown`, `display_name`, `raw_name`, `status` IN (`active`,`archived`) default `active`, `deleted_at`, audit columns). One row per real party (person/company) per tenant; roles are an attribute, not a partition.
- `actor_keys` (`id`, `actor_id`, `tenant_id`, `key_type` IN (`TAX_ID`,`LEGAL_NAME`,`PERSON_NAME`,`ALIAS`), `key_value`, `active`, `source` IN (`direct`,`import`,`backfill`)). Deduplicating keys live here.
- `actor_roles` (`actor_id`, `role` IN (`customer`,`provider`,`investor`,`manager`,`contractor`,`biller`,`lessee`), PK `(actor_id, role)`).
- Unique partial index `uq_actor_keys_active (tenant_id, key_type, key_value) WHERE active` — uniqueness lives in the index, not a trigger; only active keys count, so archived identities free their keys.
- GIN trigram indexes on `actor_keys.key_value`: name keys (`LEGAL_NAME`/`PERSON_NAME`/`ALIAS`) in migration 241; `TAX_ID` in migration 247.
- Carrier FKs `*_actor_id` (migration 242): `customers.actor_id`, `investors.actor_id`, `managers.actor_id`, `providers.actor_id`, `workorders.contractor_actor_id`, `labors.contractor_actor_id`, `invoices.biller_actor_id`, all `ON DELETE SET NULL`.

## Owned APIs

- `GET /api/v1/version`
- `GET /api/v1/healthz`
- `GET /api/v1/ping`
- `GET /api/v1/admin/tenants`
- `POST /api/v1/admin/tenants`
- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users`
- `POST /api/v1/admin/memberships`

### Actors (Identity Registry)

- `POST /api/v1/actors` — resolve-or-create. Body `{name, tax_id?, role, allow_create, reject_existing}`. 200 if reused, 201 if created.
- `GET /api/v1/actors` — list. `?status=active|archived|all&page&per_page` (default status `active`, per_page 100).
- `GET /api/v1/actors/search` — exact + similar (trigram). `?q&field=name|tax_id&limit` (default/limit cap 50).
- `GET /api/v1/actors/by-tax-id` — `?tax_id`. 200 / 404 / 422.
- `GET /api/v1/actors/similar` — advisory candidates (always 200). `?name&limit`.
- `GET /api/v1/actors/:actor_id`.
- `PUT /api/v1/actors/:actor_id` — update. Body `{display_name, party_type}`; rotates the name key when the canonical name changes (409 if another active identity already uses it).
- `DELETE /api/v1/actors/:actor_id` — hard delete (carriers' `*_actor_id` set to NULL via FK).
- `POST /api/v1/actors/:actor_id/archive` — soft delete; deactivates keys (frees them from the dedup pool).
- `POST /api/v1/actors/:actor_id/restore` — reactivates actor + keys (409 if an active identity now uses a key).
- `PUT /api/v1/actors/:actor_id/roles` — `{roles:[]}`, replaces the role set.
- `PUT /api/v1/actors/:actor_id/tax-id` — `{tax_id}`, re-keys TAX_ID; 409 on collision, 400 if non-numeric. `actor_id` does not change, so carriers stay attached.

### Registry (aggregated read-model)

- `GET /api/v1/registry` — unified search. `?q&type&status&page&per_page` (per_page 100). Response `{data:[{entity_type,id,name,tax,roles,archived}], page_info{page,per_page,total,max_page}}`. Built by `UNION ALL` over actors (once) plus the four catalogs.
- `PUT /api/v1/registry/actors/:actor_id/aliases` — `{aliases:[]}`, replaces the ALIAS key set (409 if an alias collides with another active identity).

## Dependencies On Other Domains

- None verified.

## Inbound Dependencies

- All protected API domains depend on this domain for validation/auth context.
- Business Insights depends on `OrgID` from this domain.

## Outbound Dependencies

- Google Identity Platform/JWKS.
- Firebase/Identity Platform Admin.
- PostgreSQL.

## Aggregate Roots

- `User`
- `AuthTenant`
- `AuthMembership`
- `Role`
- `Permission`
- `Actor`

## Critical Business Rules

- `GET`, `HEAD`, `OPTIONS` map to `api.read`.
- Other methods map to `api.write`.
- `admin` and `manager` receive read/write.
- `viewer` receives read only.
- `/admin/*` requires role `admin`.
- Local dev auth assigns role `admin`.

### Fine-Grained Permissions (Pilar 2)

- Fine permissions `users:manage` and `invites:write` are seeded (migration 246), mapped to `admin` and to `admin`+`tenant_owner` respectively, so access is unchanged versus the prior hardcoded checks.
- `authz.HasPermissionOrRole(ctx, permission, fallbackRoles...)` and `RequirePermissionOrRole` implement a transition dual-check: evaluate the fine permission and, if the current role lacks it, fall back to the role and log `authz fallback_to_coarse`. Once `fallback_to_coarse=0` the role grant can be retired.
- `internal/admin/handler.go` (`/admin/users`) requires `users:manage` (fallback `admin`); `internal/admin/invites.go` `canManageTenant` requires `invites:write` (fallback `admin`,`tenant_owner`). This replaced the hardcoded `role == "admin"` checks.

### Identity Resolution And Dedup (Pilar 3, Identity Gate)

- One identity per real party per tenant. Dedup runs by a cascade: TAX_ID first, then the legal/person name. The active partial index guarantees uniqueness per `(tenant_id, key_type, key_value)` with no role partition, so the same CUIT/name across roles (e.g. customer + provider) is one identity.
- `TAX_ID` is CUIT/CUIL/DNI normalized to digits only (`NormalizeTaxID`). The tax id must be numeric (`TaxIDIsNumeric`, validated only at write points — create and re-key); 400/422 otherwise. `ValidCUIT` mod-11 checksum is advisory (not enforced; foreign ids may not validate).
- Name keys: `LEGAL_NAME` for orgs, encoded `core|FORM` (e.g. `acme|SA`); `PERSON_NAME` for persons/unknown. `ParseLegalName` detects the legal form from trailing tokens; `Canonicalize` lowercases, strips vowel accents, preserves the `ñ` and word spaces (`la plata|SA` differs from `laplata|SA`). `ALIAS` keys hold alternative names.
- Strict create: `reject_existing=true` runs `LookupIdentity` first and returns 409 if the name or CUIT already exists, without reusing. Direct creates/comboboxes without the flag reuse the existing identity (normal dedup).
- Rename uniqueness: `UpdateActor` rotates the name key when the canonical name changes; 409 if the new name is used by another active identity. Changing only `party_type` (same canonical) does not touch keys.
- Tax-id re-key: `PUT /tax-id` rotates the active `TAX_ID` key in place; `actor_id` is unchanged so all carriers (`*_actor_id` FKs) stay attached. 409 if another active identity holds that CUIT (the merge case is out of scope); 400 if non-numeric. Idempotent.
- Alias edit (`PUT /registry/actors/:id/aliases`) replaces the ALIAS set; 409 on collision with another active identity; `actor_id` unchanged.
- Archive deactivates the actor's keys (removes them from the dedup pool, so a new create with the same CUIT/name makes a fresh identity); restore reactivates them and is rejected with 409 if a key is now taken.
- `StampActor` resolves-and-stamps a carrier's `*_actor_id` inside the caller transaction; it is a no-op when `IdentityGateEnabled()` is off or the name is empty. Resolution runs in the caller's transaction, so a unique-violation (race or duplicate) rolls back and the caller returns 409.

### Registry Read-Model

- `GET /registry` is read-only and tenant-scoped (flag-gated). It UNIONs actors (once, deduplicated across roles) with the four catalogs (`crops`, `types`, `lease-types`, `campaigns`).
- `type` accepts `all`, the seven actor roles (`customer`,`provider`,`investor`,`manager`,`contractor`,`biller`,`lessee`), and the four catalogs; a role filter restricts actors via `actor_roles`.
- `q` matches `actors.display_name`, active actor keys (name keys ILIKE; `TAX_ID` by digit prefix), and catalog `name` ILIKE.
- `status` filters `active` (default) / `archived` / `all` by `deleted_at`.

## Tenant Isolation Requirements

- Membership resolution injects tenant UUID into context.
- Physical tenant scoping is verified for Business Insights.
- `actors` and `actor_keys` carry a `tenant_id`. The resolver (`resolveTenant`/`TenantFor`) always fills a concrete tenant — `OrgIDFromContext` or the `default` tenant — so `actor_keys.tenant_id` is never NULL and the active unique index applies (no COALESCE, no hardcoded default UUID).
- Identity resolution, actor reads/writes, and the Registry are scoped to that tenant; cross-tenant access returns 404. Registry/alias scoping additionally gates on `TenantEnforcementEnabled()`.
- Pilar 1: `tenant_id` is `NOT NULL` with a `default` tenant DEFAULT on the eleven per-tenant tables (`customers`, `campaigns`, `projects`, `managers`, `investors`, `providers`, `crops`, `types`, `lease_types`, `business_parameters`, `categories`) per migration 245; stamping/scoping is flag-gated by `OrgIDFromContext`. RLS is out of scope.

## Security Requirements

- API key validation.
- Bearer JWT when identity auth is enabled.
- Role/permission authorization, with the fine-permission dual-check on the admin surface.
- Admin role for admin endpoints.
- Identity gating controlled by `IdentityGateEnabled()` and tenant enforcement by `TenantEnforcementEnabled()` (env-backed, `sync.Once`).
- Actor/registry endpoints mount on the validation middleware group; the tax-id numeric rule is enforced at all write points.

## Evidence

- `cmd/api/http_server.go`
- `cmd/config/auth.go`
- `internal/platform/http/middlewares/gin/middleares.go`
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/platform/http/middlewares/gin/local_dev_authz.go`
- `internal/admin/handler.go`
- `internal/admin/invites.go`
- `internal/admin/idp/*`
- `internal/shared/authz/authz.go`
- `internal/shared/models/base.go`
- `internal/actors/handler.go`
- `internal/actors/handler_crudar.go`
- `internal/actors/repository.go`
- `internal/actors/repository_crudar.go`
- `internal/identity/resolver.go`
- `internal/identity/taxid.go`
- `internal/identity/legalname.go`
- `internal/identity/canonical.go`
- `internal/registry/handler.go`
- `internal/registry/repository.go`
- `wire/wire.go`
- `migrations_v4/000180_authn_authz_mvp.up.sql`
- `migrations_v4/000201_auth_uuid_migration.up.sql`
- `migrations_v4/000241_actors_registry.up.sql`
- `migrations_v4/000242_actor_id_fks.up.sql`
- `migrations_v4/000243_actor_tenant_not_null.up.sql`
- `migrations_v4/000245_tenant_id_default_not_null.up.sql`
- `migrations_v4/000246_fine_permissions.up.sql`
- `migrations_v4/000247_actor_taxid_trgm.up.sql`
