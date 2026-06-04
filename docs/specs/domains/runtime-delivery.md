# Runtime, Migration, And Delivery Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Define verified runtime packaging, local setup, migration execution, CI/CD, deployment, rollback, and DB reset support.

## Boundaries

Owns:
- Runtime entrypoints.
- Docker/Docker Compose runtime setup.
- Migration commands and migration source.
- GitHub Actions build/test/deploy workflows.
- DB reset/deploy support scripts where verified.

Does not own:
- Business behavior.
- External infrastructure code not present in this repository.
- Monitoring/on-call procedures.

## Owned Entities

No business entities.

Operational artifacts:
- `cmd/api/*`
- `cmd/migrate/*`
- `migrations_v4/*`
- `.github/workflows/*`
- `Dockerfile`
- `Dockerfile.dev`
- `docker-compose.yml`
- `Makefile`
- `scripts/*`

## Owned APIs

None beyond health/version surfaces covered by Platform, Identity, And Admin.

## Dependencies On Other Domains

- All application domains depend on runtime and migrations to run.

## Inbound Dependencies

- Developers.
- CI/CD workflows.
- Deployment operators.

## Outbound Dependencies

- Docker.
- GitHub Actions.
- GCP/Cloud Run/Artifact Registry/Secret Manager/Workload Identity as referenced by workflows.
- PostgreSQL.

## Aggregate Roots

None.

## Critical Business Rules

- Migrations under `migrations_v4/` are ordered SQL migration source.
- CI workflows include PR CI and environment deployments.
- Rollback workflows exist for staging and production.
- DB reset workflows exist for dev/staging from production.

UNKNOWN:
- Full production infrastructure topology.
- Manual release approval policy beyond workflow evidence.
- Incident escalation process.

## Tenant Isolation Requirements

N/A directly. Runtime must preserve Platform, Identity, And Admin tenant context and DB configuration, but no separate runtime tenant model is verified.

## Security Requirements

- Secrets/configuration are environment/workflow driven.
- Exact production secret inventory is UNKNOWN.

## Evidence

- `cmd/api/main.go`
- `cmd/api/http_server.go`
- `cmd/migrate/*`
- `cmd/config/*`
- `Dockerfile`
- `Dockerfile.dev`
- `docker-compose.yml`
- `Makefile`
- `.github/workflows/*`
- `scripts/*`
