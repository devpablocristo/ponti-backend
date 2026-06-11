# External Services Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Ponti AI / Axis

Verified config/code:
- `AI_PROVIDER`
- `AI_AXIS_ENABLED`
- `AXIS_COMPANION_BASE_URL`
- `AXIS_COMPANION_API_KEY`
- `AXIS_PRODUCT_SURFACE`
- `PONTI_AXIS_API_KEY`
- `AI_SERVICE_URL`
- `AI_SERVICE_KEY`
- AI timeout config for legacy fallback.

Behavior:
- With `AI_PROVIDER=axis` and `AI_AXIS_ENABLED=true`, Ponti backend calls
  Axis Companion and adapts the response to the legacy web contract.
- Axis 4xx responses are returned closed. Network errors and Axis 5xx fallback
  to the legacy `AI_SERVICE_*` path.
- If neither Axis nor legacy AI is configured, chat falls back to dummy
  responses and stream emits `ai_not_configured`.
- Product integration endpoints accept `Authorization: Bearer <PONTI_AXIS_API_KEY>`
  only on the Axis/Ponti capability surface.

Evidence:
- `cmd/config/ai.go`
- `internal/axis/client.go`
- `internal/ai/capabilities.go`
- `internal/ai/client.go`
- `internal/ai/usecases/usecases.go`
- `internal/platform/http/middlewares/gin/axis_product_integration_auth.go`

UNKNOWN:
- Production Nexus policy ownership for `agent.capability.invoke`.
- Production Axis Companion rollout/rollback owner.

## Review / Nexus

Verified config/code:
- Review URL/API key config.
- Optional client setup.
- Negative stock insight policy call.

Evidence:
- `cmd/config/review.go`
- `internal/reviewproxy/client.go`
- `internal/businessinsights/service.go`
- `cmd/api/http_server.go`

UNKNOWN:
- Review/Nexus ownership.
- Review/Nexus health checks.
- Policy deployment/release process.

## Identity Platform

Verified:
- JWT/JWKS auth.
- Admin user provisioning adapter.

Evidence:
- `cmd/config/auth.go`
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/admin/idp/firebase_admin.go`

UNKNOWN:
- Identity provider incident response.
- Key rotation process.
- Tenant provisioning process outside admin API.
