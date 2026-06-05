# Backup And Restore Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Verified Artifacts

- Existing backup runbook file exists outside SDD specs: `docs/PROD_DB_BACKUP_RUNBOOK.md`.
- DB reset scripts exist.
- Dev/staging reset-from-production workflows exist.
- Post-restore hardening script exists.

Evidence:
- `docs/PROD_DB_BACKUP_RUNBOOK.md`
- `scripts/db/reset-local-db-from-prod.sh`
- `scripts/db/hardening_post_restore.sql`
- `.github/workflows/reset-dev-db-from-prod.yml`
- `.github/workflows/reset-stg-db-from-prod.yml`

## UNKNOWN

- Backup schedule.
- Backup retention policy.
- Restore RTO/RPO.
- Production restore approval process.
- Whether backups are automatically tested.
- Whether restore drills are scheduled.

## Constraint

This baseline spec does not modify or replace the existing historical runbook. It only records verified evidence.

The detailed operational spec for local reset from production data lives at `../scripts/reset-local-db-from-prod/spec.md`. That script-specific spec is authoritative only for that local reset workflow.
