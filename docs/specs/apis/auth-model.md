# API Auth Model Baseline Specification

Specification type: baseline current-state API security specification.

## Baseline Validation

Protected API routes use validation middleware from the platform HTTP middleware setup.

Verified mechanisms:
- API key validation.
- Identity Platform authz when enabled.
- Local dev authz when auth is disabled.

Evidence:
- `internal/platform/http/middlewares/gin/middleares.go`
- `cmd/config/auth.go`

## Permission Mapping

| HTTP method | Required permission |
|---|---|
| `GET` | `api.read` |
| `HEAD` | `api.read` |
| `OPTIONS` | `api.read` |
| All other methods | `api.write` |

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`

## Roles

Verified seeded roles:
- `admin`
- `manager`
- `viewer`

Verified permissions:
- `admin`: `api.read`, `api.write`
- `manager`: `api.read`, `api.write`
- `viewer`: `api.read`

Evidence:
- `migrations_v4/000180_authn_authz_mvp.up.sql`

## Admin Routes

Admin endpoints require role `admin`.

Evidence:
- `internal/admin/handler.go`

## Tenant Context

Verified request context keys:
- Actor
- OrgID
- Role
- Scopes

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/platform/http/middlewares/gin/local_dev_authz.go`
- `internal/shared/handlers/auth.go`

UNKNOWN:
- Whether every business API verifies tenant membership against workspace ownership.
- Any auth behavior enforced by external frontend/BFF.
