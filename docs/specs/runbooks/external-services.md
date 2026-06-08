# External Services Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Ponti AI

Verified config/code:
- `AI_SERVICE_URL`
- `AI_SERVICE_KEY`
- AI timeout config.

Behavior:
- If AI is not configured, chat falls back to dummy responses and stream emits `ai_not_configured`.

Evidence:
- `cmd/config/ai.go`
- `internal/ai/client.go`
- `internal/ai/usecases/usecases.go`

UNKNOWN:
- AI service deployment.
- AI service health checks.
- AI service incident procedure.

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
