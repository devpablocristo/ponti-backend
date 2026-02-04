# Despliegue de ponti-backend en Google Cloud Run

## TLDR
- **DEV**: push a `develop` → deploy automático.
- **STG**: push a `main` → deploy automático con **tag SHA**.
- **PROD**: promoción manual del **mismo artefacto** probado en STG.
- **Preview**: on‑demand por PR (DB efímera).

## Workflows

- `ci-pr.yml`: PR a `develop` → tests.
- `deploy-dev.yml`: push a `develop` → DEV.
- `deploy-staging.yml`: push a `main` → STG.
- `promote-prod.yml`: manual → PROD (usa SHA de STG).
- `deploy-preview.yml`: manual o label `preview`.
- `cleanup-preview.yml`: cleanup al cerrar PR + cron semanal.
- `reset-dev.yml`: reset de DEV (golden snapshot).
- `refresh-golden-snapshot.yml`: genera snapshot desde STG (DB `new_ponti_db_staging` en instancia `new-ponti-db-dev`).

## Environments

En GitHub: `dev`, `stg`, `prod`.  

## Reset DEV (pasos concretos)

Workflow: `reset-dev.yml` (manual).

1. GitHub → Actions → **Reset DEV**.
2. Run workflow.
3. Verificar que el workflow termine OK.
4. (Opcional) Probar health: `https://<dev-backend>/ping`.

Qué hace:
- Borra `new_ponti_db_dev`.
- Restaura el **Golden Snapshot**.
- Ejecuta hardening (si `HARDENING_SQL_URI` está configurado).
- Corre smoke test (si `SMOKE_TEST_URL` está configurado).

## Golden Snapshot (pasos concretos)

Workflow: `refresh-golden-snapshot.yml` (manual o cron).

1. GitHub → Actions → **Refresh Golden Snapshot**.
2. Run workflow.
3. Verificar que el snapshot se haya exportado al bucket definido en `GOLDEN_SNAPSHOT_BUCKET`.

El snapshot se usa luego por `reset-dev.yml`.

**Requisito:** El SA `github-actions@new-ponti-stg` debe tener `roles/cloudsql.admin` (o equivalente) en el proyecto **new-ponti-dev** para poder exportar. Ver [GITHUB_SECRETS.md](GITHUB_SECRETS.md#permiso-iam-pendiente-refresh-golden-snapshot).
