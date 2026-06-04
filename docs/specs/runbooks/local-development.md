# Local Development Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Verified Local Artifacts

- `docker-compose.yml`
- `Dockerfile.dev`
- `Makefile`
- `scripts/run_ponti_local.sh`
- `scripts/down_ponti_local.sh`
- `scripts/db/db_reset.sh`
- `scripts/db/db_migrate_up.sh`
- `cmd/config/*`

## Verified Capabilities

The repository contains local development support for:

- Docker-based local runtime.
- Local database reset/migration scripts.
- Environment-driven config loading.

Evidence:
- `docker-compose.yml`
- `Dockerfile.dev`
- `scripts/*`
- `cmd/config/loadconfig.go`
- `cmd/config/db.go`

## UNKNOWN

- Required local environment variables as a complete list.
- Required external service availability for full local behavior.
- Canonical local startup command if multiple scripts/Make targets exist.
- Local seed data policy.

## Operational Constraint

Local dev auth may assign admin role when `AUTH_ENABLED=false`.

Evidence:
- `internal/platform/http/middlewares/gin/local_dev_authz.go`

## Script-Specific Spec Boundary

The detailed operational spec for `make reset-local-db-from-prod` lives at `../scripts/reset-local-db-from-prod/spec.md`. This runbook only summarizes verified local development capabilities.
