# Deployment Runbook Baseline Specification

Specification type: baseline current-state operational specification.

## Verified CI/CD Workflows

Workflow files verified:

- `ci-pr.yml`
- `deploy-dev.yml`
- `deploy-staging.yml`
- `deploy-prod.yml`
- `approve-staging.yml`
- `release.yml`
- `rollback-staging.yml`
- `rollback-prod.yml`
- `audit-service-alignment.yml`

Evidence:
- `.github/workflows/*`

## Verified Runtime Packaging

- Dockerfile exists for backend container build.
- Dockerfile.dev exists for development runtime.

Evidence:
- `Dockerfile`
- `Dockerfile.dev`

## Verified Deployment Targets

Workflows reference GCP/Cloud Run oriented deployment behavior.

Evidence:
- `.github/workflows/deploy-dev.yml`
- `.github/workflows/deploy-staging.yml`
- `.github/workflows/deploy-prod.yml`

## Rollback

Rollback workflows exist for staging and production.

Evidence:
- `.github/workflows/rollback-staging.yml`
- `.github/workflows/rollback-prod.yml`

UNKNOWN:
- Human release manager.
- Exact approval policy beyond workflow files.
- Production runtime topology.
- Blue/green or canary policy.
- Rollback validation checklist.
