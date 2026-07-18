# External Contracts Baseline Specification

Specification type: baseline current-state external API specification.

## Ponti AI

Ponti backend preserves the web-facing AI contract while routing internally to
either legacy `ponti-ai` or Axis.

Legacy backend proxy routes:
- `POST /api/v1/ai/chat` -> legacy `POST /v1/chat`
- `POST /api/v1/ai/chat/stream` -> legacy `POST /v1/chat/stream`
- `GET /api/v1/ai/chat/conversations` -> legacy `GET /v1/chat/conversations?limit=n`
- `GET /api/v1/ai/chat/conversations/:conversation_id` -> legacy `GET /v1/chat/conversations/:conversation_id`

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
- Axis 4xx responses do not fallback because they normally represent auth,
  tenant, policy, installation or contract failures.

Web response fields preserved by Ponti:
- `request_id`
- `output_kind`
- `content_language`
- `chat_id`
- `reply`
- `tokens_used`
- `tool_calls`
- `pending_confirmations`
- `blocks`
- `routed_agent`
- `routing_source`

Optional Axis diagnostic fields:
- `axis_run_id`
- `axis_task_id`
- `run_id`
- `task_id`

UNKNOWN:
- AI service storage and auth internals.

## Axis Companion

Backend routes preserving Ponti web contract:
- `POST /api/v1/ai/chat` -> Axis `POST /v1/chat`
- `GET /api/v1/ai/chat/conversations` -> Axis `GET /v1/chat/conversations?limit=n`
- `GET /api/v1/ai/chat/conversations/:conversation_id` -> Axis `GET /v1/chat/conversations/:conversation_id`

Canonical product mapping:
- `product_surface=ponti`
- Axis `org_id=auth_tenants.id`
- `external_tenant_id=auth_tenants.id` initially
- workspace v1: `{ customer_id, project_id, campaign_id, field_id }`

Ponti product metadata/capabilities exposed to Axis:
- `GET /api/v1/capabilities`
- `GET /api/v1/insights`
- `GET /api/v1/insights/summary`
- `GET /api/v1/insights/:id/explain`
- `POST /api/v1/ai/actions/insight-resolve/prepare`
- `POST /api/v1/ai/actions/workorder-draft/prepare`
- `POST /api/v1/ai/actions/stock-adjustment/prepare`

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
- First rollout is read-only plus governed previews. The three `prepare`
  endpoints return previews and must not mutate Ponti data.
- Sensitive writes are governed by Nexus with
  `nexus_action_type=agent.capability.invoke`.

Local operational smoke:
- `make smoke-axis-all` runs onboarding, read-only execution, draft governance,
  direct previews, Nexus-approved preview execution, and chat.
- `smoke-ponti-axis-nexus-approved-draft.sh` creates/updates the local Nexus
  action type and policy idempotently, approves one request, executes the
  preview capability through Axis and reports the Nexus request as `executed`.

## Review / Nexus Governance

Purpose:
- Evaluate whether reactive domain events should create business insight
  candidates.
- Govern Ponti draft/preview capabilities invoked through Axis.

Verified action type:
- `ponti.stock.negative`
- `agent.capability.invoke`

Verified requester:
- `ponti-backend`
- `ponti-nexus-smoke-requester` in local smoke

Evidence:
- `internal/businessinsights/service.go`
- `internal/reviewproxy/client.go`
- `cmd/config/review.go`
- `scripts/axis/smoke-ponti-axis-nexus-approved-draft.sh`

UNKNOWN:
- Production Nexus policy configuration.

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
