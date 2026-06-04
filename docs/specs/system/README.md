# System Baseline Specifications

Specification type: baseline current-state specifications.

These files describe the verified current system architecture. They are not generic documentation and they are not future-state design proposals.

## Files

- `overview.md`: runtime shape, business purpose, modules, and verified architecture.
- `security-and-tenancy.md`: API key, JWT/local auth, roles, permissions, and tenant isolation evidence.
- `external-systems.md`: verified external services and unavailable external code.
- `observability.md`: verified logging/observability mechanisms and UNKNOWN gaps.

## Evidence Rules

- Claims must be backed by current repository evidence.
- Missing evidence is marked `UNKNOWN`.
- Historical or planned features are not treated as implemented unless current code/config/migrations verify them.

## Primary Evidence

- `cmd/api/http_server.go`
- `cmd/api/main.go`
- `cmd/config/*`
- `internal/*/handler.go`
- `internal/*/usecases*.go`
- `internal/platform/http/middlewares/gin/*`
- `migrations_v4/*`
- `.github/workflows/*`
- `Dockerfile`
- `docker-compose.yml`
- `Makefile`
