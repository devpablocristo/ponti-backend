# Estado actual de workflows

## TLDR

- `develop` → deploy automático a DEV.
- STG, PROD y Preview: **no implementados** en este repo.
- Reset DEV y Golden Snapshot: manual / cron.

## Workflows implementados

### 1) CI PR
Archivo: `.github/workflows/ci-pr.yml`

**Triggers:** `pull_request` a `develop`

**Jobs:** lint, build, test, security-scan (govulncheck, no bloquea)

### 2) Deploy DEV
Archivo: `.github/workflows/deploy-dev.yml`

**Triggers:** `push` a `develop`

**Comportamiento:**
- Build de imagen Docker.
- Push a Artifact Registry.
- Deploy a Cloud Run en proyecto DEV.
- DB: `new_ponti_db_dev`, instancia `new-ponti-db-dev`.

### 3) Reset DEV
Archivo: `.github/workflows/reset-dev.yml`

**Triggers:** `workflow_dispatch` (manual)

**Comportamiento:**
- Borra DB dev.
- Crea DB vacía.
- Restaura Golden Snapshot.
- Opcional: hardening, smoke test.

### 4) Refresh Golden Snapshot
Archivo: `.github/workflows/refresh-golden-snapshot.yml`

**Triggers:** `workflow_dispatch` o cron (lunes 4:00 UTC)

**Comportamiento:**
- Exporta DB de STG (`new_ponti_db_staging`) a GCS.
- Copia a `latest.sql.gz` para uso de reset-dev.

## Infraestructura

| Ambiente | Proyecto | Cloud SQL | DB |
|----------|----------|-----------|-----|
| DEV | new-ponti-dev | new-ponti-db-dev | new_ponti_db_dev |
| STG | new-ponti-stg | new-ponti-db-dev (cross-project) | new_ponti_db_staging |
| PROD | new-ponti-prod | new-ponti-prod-db | ponti_api_db |
