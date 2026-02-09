# Despliegue de ponti-backend en Google Cloud Run

## TLDR

- **DEV**: push a `develop` → deploy automático.
- **STG / PROD / Preview**: workflows no implementados actualmente.
- **Reset DEV**: manual, restaura Golden Snapshot.
- **Golden Snapshot**: manual o cron (lunes 4am), exporta desde STG.

## Workflows actuales

| Archivo | Trigger | Descripción |
|---------|---------|-------------|
| `ci-pr.yml` | PR a `develop` | Lint, build, test, govulncheck |
| `deploy-dev.yml` | Push a `develop` | Deploy a Cloud Run DEV |
| `reset-dev.yml` | Manual | Borra DB dev, restaura Golden Snapshot |
| `refresh-golden-snapshot.yml` | Manual o cron (lun 4am) | Exporta snapshot desde STG al bucket |

## Reset DEV (pasos concretos)

Workflow: `reset-dev.yml` (manual).

1. GitHub → Actions → **Reset DEV**.
2. Run workflow.
3. Verificar que termine OK.
4. (Opcional) Probar: `https://<dev-backend>/ping`.

Qué hace:
- Borra `new_ponti_db_dev`.
- Crea la DB vacía.
- Restaura el **Golden Snapshot** desde `GOLDEN_SNAPSHOT_URI`.
- Ejecuta hardening (si `HARDENING_SQL_URI` está configurado).
- Smoke test (si `SMOKE_TEST_URL` está configurado).

## Golden Snapshot (pasos concretos)

Workflow: `refresh-golden-snapshot.yml` (manual o cron).

1. GitHub → Actions → **Refresh Golden Snapshot**.
2. Run workflow.
3. Verificar que el snapshot se exporte al bucket (`GOLDEN_SNAPSHOT_BUCKET`).

El snapshot se usa luego por `reset-dev.yml`.

**Requisito:** El SA de GitHub Actions debe tener permisos para exportar desde la instancia Cloud SQL de STG. Ver [GITHUB_SECRETS.md](GITHUB_SECRETS.md#permiso-iam-pendiente-refresh-golden-snapshot).
