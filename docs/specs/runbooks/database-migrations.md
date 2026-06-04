# Database Migrations Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Verified Migration Source

- SQL migrations live under `migrations_v4/`.
- Migration command code exists under `cmd/migrate/`.
- DB migration scripts exist under `scripts/db/`.

Evidence:
- `migrations_v4/*`
- `cmd/migrate/main.go`
- `cmd/migrate/migrate_sql.go`
- `scripts/db/db_migrate_up.sh`
- `scripts/db/db_validate.sh`
- `Makefile`

## Verified Migration-Related Workflows

- Production migration workflow exists.
- Database reset workflows exist for dev/staging from production.

Evidence:
- `.github/workflows/apply-migrations-prod.yml`
- `.github/workflows/reset-dev-db-from-prod.yml`
- `.github/workflows/reset-stg-db-from-prod.yml`

## UNKNOWN

- Manual production migration procedure outside workflows.
- Roll-forward-only policy.
- Migration approval owners.
- Database backup requirement before migrations.
- Production lock/maintenance window policy.

## Script-Specific Spec Boundary

The detailed operational spec for local reset from production data lives at `../scripts/reset-local-db-from-prod/spec.md`. This runbook only summarizes verified migration sources and migration-related workflows.
