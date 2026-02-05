# Documentación ponti-backend

## Índice

| Doc | Descripción |
|-----|-------------|
| [DEPLOY.md](DEPLOY.md) | Workflows de deploy, reset, golden snapshot |
| [ESTADO_FINAL_WORKFLOWS.md](ESTADO_FINAL_WORKFLOWS.md) | Detalle de workflows y comportamiento actual |
| [CONFIGURAR_VARIABLES_GITHUB.md](CONFIGURAR_VARIABLES_GITHUB.md) | Variables y secrets para GitHub Actions |
| [GITHUB_SECRETS.md](GITHUB_SECRETS.md) | Secrets, WIF e IAM (refresh-golden-snapshot) |
| [GCP_DB_CREDS.md](GCP_DB_CREDS.md) | Credenciales Cloud SQL (dev) |
| [DIAGNOSTICO_CLOUD_RUN.md](DIAGNOSTICO_CLOUD_RUN.md) | Troubleshooting Cloud Run |
| [SETUP_PROD.md](SETUP_PROD.md) | Setup proyecto producción |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Arquitectura del backend |
| [FEATURE-MAP.md](FEATURE-MAP.md) | Mapa de features y vistas |

## Infraestructura (estado actual)

| Ambiente | Proyecto | Cloud SQL | DB |
|----------|----------|-----------|-----|
| DEV | new-ponti-dev | new-ponti-db-dev | new_ponti_db_dev |
| STG | new-ponti-stg | new-ponti-db-dev (cross-project) | new_ponti_db_staging |
| PROD | new-ponti-prod | new-ponti-prod-db | ponti_api_db |

## Histórico / referencia

| Doc | Contenido |
|-----|-----------|
| [UNIFICAR_DEV_STG_CLOUDSQL.md](UNIFICAR_DEV_STG_CLOUDSQL.md) | Plan de unificación DEV/STG (ejecutado) |
| [RUNBOOK_UNIFICAR_STG.md](RUNBOOK_UNIFICAR_STG.md) | Runbook unificación |
| [UNIFICACION_STG_VERIFICACION.md](UNIFICACION_STG_VERIFICACION.md) | Verificación post-unificación |
| [RENAME_DB_DEV.md](RENAME_DB_DEV.md) | Rename ponti_api_db → new_ponti_db_dev |
| [RESUMEN_APAGADO_STG_Y_AUDITORIA.md](RESUMEN_APAGADO_STG_Y_AUDITORIA.md) | Apagado instancia stg, auditoría costos |
