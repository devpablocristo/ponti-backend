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

UNKNOWN:
- External request/response schema beyond proxy behavior.
- AI service storage and auth internals.

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
