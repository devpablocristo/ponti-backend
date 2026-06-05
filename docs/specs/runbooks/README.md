# Runbook Baseline Specifications

Specification type: baseline current-state operational specifications.

These files describe verified operational behavior only. Missing operational evidence is marked `UNKNOWN`.

Runbooks summarize operational areas. Script-specific specs under `docs/specs/scripts/` are authoritative only for the named script they describe.

## Files

- `local-development.md`
- `database-migrations.md`
- `deployment.md`
- `backup-restore.md`
- `external-services.md`
- `incident-response.md`

## Script-Specific Specs

- `../scripts/reset-local-db-from-prod/spec.md`: detailed operational spec for `scripts/db/reset-local-db-from-prod.sh` / `make reset-local-db-from-prod`.

## Evidence

- `Makefile`
- `Dockerfile`
- `Dockerfile.dev`
- `docker-compose.yml`
- `.github/workflows/*`
- `scripts/*`
- `cmd/config/*`
- `cmd/migrate/*`
