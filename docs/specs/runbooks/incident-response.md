# Incident Response Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Verified Signals

- Health endpoints exist.
- Request/response logging middleware exists.
- Authz decision logging exists.
- Rollback workflows exist for staging and production.

Evidence:
- `cmd/api/http_server.go`
- `internal/platform/http/middlewares/gin/request_and_response_logger.go`
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `.github/workflows/rollback-staging.yml`
- `.github/workflows/rollback-prod.yml`

## UNKNOWN

- Incident severity definitions.
- On-call rota.
- Alert channels.
- Monitoring dashboards.
- Metrics and tracing instrumentation.
- Error budget/SLOs.
- Customer communication process.
- Postmortem process.

## Baseline Constraint

This spec must not invent an incident procedure. Future operational specs must fill UNKNOWNs only when backed by verified runbook/config/workflow evidence.
