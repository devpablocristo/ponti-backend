# Security And Tenancy Baseline Specification

Specification type: baseline current-state specification.

## Authentication

Verified authentication mechanisms:

- API key validation is part of the API validation middleware.
- When identity auth is enabled, requests require a bearer JWT verified through JWKS.
- When local dev auth is enabled, the middleware accepts local/fake identity context and assigns admin role.

Evidence:
- `internal/platform/http/middlewares/gin/middleares.go`
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/platform/http/middlewares/gin/local_dev_authz.go`
- `cmd/config/auth.go`

## Authorization

Verified permission model:

- `GET`, `HEAD`, and `OPTIONS` require `api.read`.
- Other HTTP methods require `api.write`.
- Seeded roles are `admin`, `manager`, and `viewer`.
- `admin` and `manager` receive read/write permissions.
- `viewer` receives read permission.
- `/api/v1/admin/*` requires role `admin` in addition to baseline validation.

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/admin/handler.go`
- `migrations_v4/000180_authn_authz_mvp.up.sql`
- `migrations_v4/000201_auth_uuid_migration.up.sql`

## Tenant Context

Verified tenant context:

- Identity middleware resolves membership and injects `OrgID` into request context.
- Local dev middleware attempts to resolve tenant from header or default tenant.
- Business insights use `OrgID` as tenant id for tenant-scoped candidate operations.

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/platform/http/middlewares/gin/local_dev_authz.go`
- `internal/shared/handlers/auth.go`
- `internal/businessinsights/handler.go`
- `internal/businessinsights/repository.go`

## Tenant Isolation

Verified physical tenant isolation:

- `business_insight_candidates.tenant_id` references `auth_tenants(id)`.
- Business insight list/read/resolve/reopen queries are tenant-scoped.

Evidence:
- `migrations_v4/000209_business_insight_candidates.up.sql`
- `migrations_v4/000210_business_insight_reads.up.sql`
- `internal/businessinsights/repository.go`
- `internal/businessinsights/handler.go`

Verified workspace/project isolation:

- Shared workspace filters support `customer_id`, `project_id`, `campaign_id`, and optional `field_id`.
- Several reporting/list APIs require `customer_id`, `project_id`, and `campaign_id`.
- Many business tables are linked through `project_id` and related FKs.

Evidence:
- `internal/shared/handlers/workspace_filters.go`
- `internal/shared/filters/workspace.go`
- `internal/work-order/usecases.go`
- `internal/dashboard/handler.go`
- `internal/report/handler.go`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`

Important limitation:

- Most agricultural, operational, inventory, and finance tables do not have verified physical `tenant_id` columns.
- Therefore tenant isolation for most business data is not verified as physical row-level tenant isolation. It is currently evidenced as workspace/project-based filtering plus FK relationships.

UNKNOWN:
- Whether every API path consistently enforces workspace ownership against tenant membership.
- Whether external frontend/BFF adds additional tenant constraints.
- Whether database row-level security exists outside this repository.

## Secrets

Verified secret/config expectations:

- AI service uses service key configuration.
- Review/Nexus client uses review URL/API key configuration.
- Identity Platform/JWKS configuration is environment driven.
- GitHub/GCP workflows reference external secret handling.

Evidence:
- `cmd/config/ai.go`
- `cmd/config/review.go`
- `cmd/config/auth.go`
- `.github/workflows/*`

UNKNOWN:
- Full production secret inventory.
- Secret rotation process.
- Break-glass access process.
