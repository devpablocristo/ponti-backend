# OpenAPI Status Baseline Specification

Specification type: baseline current-state API contract status specification.

## Current Status

Status: Partially Implemented.

Verified:
- A Makefile target/reference for OpenAPI exists.
- Current REST routes are registered manually in Gin handlers.

Not verified:
- Generated OpenAPI artifacts under `docs/openapi`.
- Complete OpenAPI schema coverage for all endpoints.
- Contract publication workflow.

## Evidence

- `Makefile`
- `internal/*/handler.go`
- `cmd/api/http_server.go`

## Requirements For Future Specs

Future API contract work must treat current route inventory as baseline evidence and must not assume an OpenAPI contract exists unless generated artifacts are verified.
