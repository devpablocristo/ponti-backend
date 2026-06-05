# Platform, Identity, And Admin Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Provide API authentication, authorization, tenant context, user/tenant/membership administration, and runtime health/version surfaces.

## Boundaries

Owns:
- API validation middleware.
- API key and JWT/local auth behavior.
- Role/permission resolution.
- Admin tenant/user/membership endpoints.
- Auth tables and membership tables.

Does not own:
- Business data tenancy guarantees for most agricultural/financial tables.
- Frontend/BFF identity behavior.

## Owned Entities

- `users`
- `auth_tenants`
- `auth_roles`
- `auth_permissions`
- `auth_role_permissions`
- `auth_memberships`

## Owned APIs

- `GET /api/v1/version`
- `GET /api/v1/healthz`
- `GET /api/v1/ping`
- `GET /api/v1/admin/tenants`
- `POST /api/v1/admin/tenants`
- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users`
- `POST /api/v1/admin/memberships`

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

## Critical Business Rules

- `GET`, `HEAD`, `OPTIONS` map to `api.read`.
- Other methods map to `api.write`.
- `admin` and `manager` receive read/write.
- `viewer` receives read only.
- `/admin/*` requires role `admin`.
- Local dev auth assigns role `admin`.

## Tenant Isolation Requirements

- Membership resolution injects tenant UUID into context.
- Physical tenant scoping is verified for Business Insights.
- Physical tenant scoping is UNKNOWN for most business tables.

## Security Requirements

- API key validation.
- Bearer JWT when identity auth is enabled.
- Role/permission authorization.
- Admin role for admin endpoints.

## Evidence

- `cmd/api/http_server.go`
- `cmd/config/auth.go`
- `internal/platform/http/middlewares/gin/middleares.go`
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/platform/http/middlewares/gin/local_dev_authz.go`
- `internal/admin/handler.go`
- `internal/admin/idp/*`
- `migrations_v4/000180_authn_authz_mvp.up.sql`
- `migrations_v4/000201_auth_uuid_migration.up.sql`
