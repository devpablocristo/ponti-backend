# External Contracts Baseline Specification

Specification type: baseline current-state external API specification.

## Ponti AI

Backend proxy routes:
- `POST /api/v1/ai/chat` -> external `POST /v1/chat`
- `POST /api/v1/ai/chat/stream` -> external `POST /v1/chat/stream`
- `GET /api/v1/ai/chat/conversations` -> external `GET /v1/chat/conversations?limit=n`
- `GET /api/v1/ai/chat/conversations/:conversation_id` -> external `GET /v1/chat/conversations/:conversation_id`

Headers sent to external AI:
- `X-SERVICE-KEY`
- `X-USER-ID`
- `X-PROJECT-ID`

Evidence:
- `internal/ai/client.go`
- `internal/ai/usecases/usecases.go`
- `internal/ai/handler.go`

Status:
- Legacy provider when `AI_PROVIDER=legacy` or `AI_AXIS_ENABLED=false`.
- Temporary fallback when Axis is configured but unavailable by network/server error.

UNKNOWN:
- External request/response schema beyond proxy behavior.
- AI service storage and auth internals.

## Axis Companion

Backend routes preserving Ponti web contract:
- `POST /api/v1/ai/chat` -> Axis `POST /v1/chat`
- `GET /api/v1/ai/chat/conversations` -> Axis `GET /v1/chat/conversations?limit=n`
- `GET /api/v1/ai/chat/conversations/:conversation_id` -> Axis `GET /v1/chat/conversations/:conversation_id`

Ponti product metadata/capabilities exposed to Axis:
- `GET /api/v1/capabilities`
- `GET /api/v1/insights`
- `GET /api/v1/insights/summary`
- `GET /api/v1/insights/:id/explain`

Auth accepted on these Ponti product endpoints:
- Normal Ponti API key + user JWT flow.
- Axis product integration bearer, limited to these endpoints, when `Authorization: Bearer <value>` matches `PONTI_AXIS_API_KEY`.

Headers sent to Axis:
- `X-API-Key`
- `X-Org-ID`
- `X-User-ID`
- `X-On-Behalf-Of`
- `X-Product-Surface`
- `X-Auth-Scopes`

Evidence:
- `internal/axis/client.go`
- `internal/ai/usecases/usecases.go`
- `internal/ai/capabilities.go`
- `internal/businessinsights/handler.go`
- `docs/specs/system/ponti-ai-axis-v1.md`

Status:
- Disabled by default.
- Enabled with `AI_PROVIDER=axis` and `AI_AXIS_ENABLED=true`.

## Review / Nexus Governance

Purpose:
- Evaluate whether negative stock should create a business insight candidate.

Verified action type:
- `ponti.stock.negative`

Verified requester:
- `ponti-backend`

Evidence:
- `internal/businessinsights/service.go`
- `internal/reviewproxy/client.go`
- `cmd/config/review.go`

UNKNOWN:
- Review/Nexus policy configuration.
- Full governance API contract.

## Identity Provider

Verified behavior:
- Backend verifies bearer JWT through JWKS.
- Backend can provision/admin users through Identity Platform/Firebase admin adapter.

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/admin/idp/firebase_admin.go`

UNKNOWN:
- Identity provider tenant/project configuration outside env vars.

## Frontend / BFF

Evidence indicates an external frontend/BFF consumes backend APIs.

UNKNOWN:
- Frontend source code.
- BFF source code.
- Client-side API contract.
- User-facing workflows.
