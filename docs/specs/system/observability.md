# Observability Baseline Specification

Specification type: baseline current-state specification.

## Logging

Verified logging:

- Request/response logging middleware is registered as part of API middleware setup.
- Authz decisions are logged with user id, tenant id, route, required permission, result, timestamp, and latency.

Evidence:
- `internal/platform/http/middlewares/gin/middleares.go`
- `internal/platform/http/middlewares/gin/request_and_response_logger.go`
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `internal/platform/http/middlewares/gin/local_dev_authz.go`

## Health Checks

Verified endpoints:

- `/api/v1/healthz`
- `/api/v1/ping`
- `/api/v1/version`
- root `/healthz` and `/ping` are also registered through the platform server path.

Evidence:
- `cmd/api/http_server.go`
- `internal/platform/http/servers/gin/server.go`

## Metrics

UNKNOWN:
- No metrics exporter, metrics endpoint, Prometheus integration, OpenTelemetry metrics setup, or dashboard configuration was verified in the current audit.

## Tracing

UNKNOWN:
- No distributed tracing setup was verified in the audited code/config.

## Monitoring

UNKNOWN:
- No monitoring dashboard configuration was verified in this repository.

## Alerting

UNKNOWN:
- No alerting rules, paging configuration, or incident notification policy was verified in this repository.

## Operational Maturity Baseline

Verified:

- API health endpoints exist.
- Request/auth decision logs exist.
- CI/CD workflows exist.

UNKNOWN:

- Production monitoring maturity.
- Error budget/SLO definitions.
- On-call escalation process.
- Alert ownership.
