# API Baseline Specifications

Specification type: baseline current-state API contract specifications.

These specs describe verified API surfaces in the current backend. They are not generated OpenAPI output.

## Files

- `rest-api-inventory.md`
- `route-feature-matrix.md`
- `auth-model.md`
- `external-contracts.md`
- `non-rest-contracts.md`
- `openapi-status.md`

## Evidence

- `cmd/api/http_server.go`
- `internal/*/handler.go`
- `internal/platform/http/middlewares/gin/*`
- `internal/ai/client.go`
- `internal/reviewproxy/client.go`
