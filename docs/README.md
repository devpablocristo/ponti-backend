# Documentación ponti-backend

## Índice

| Doc | Descripción |
|-----|-------------|
| [DEPLOY.md](DEPLOY.md) | Workflows de deploy, reset, golden snapshot |
| [ESTADO_FINAL_WORKFLOWS.md](ESTADO_FINAL_WORKFLOWS.md) | Detalle de workflows actuales |
| [CONFIGURAR_VARIABLES_GITHUB.md](CONFIGURAR_VARIABLES_GITHUB.md) | Variables y secrets para GitHub Actions |
| [GITHUB_SECRETS.md](GITHUB_SECRETS.md) | Secrets, WIF e IAM (refresh-golden-snapshot) |
| [GCP_DB_CREDS.md](GCP_DB_CREDS.md) | Credenciales Cloud SQL (dev) |
| [DIAGNOSTICO_CLOUD_RUN.md](DIAGNOSTICO_CLOUD_RUN.md) | Troubleshooting Cloud Run |
| [SETUP_PROD.md](SETUP_PROD.md) | Setup proyecto producción |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Arquitectura del backend |
| [FEATURE-MAP.md](FEATURE-MAP.md) | Mapa de features y vistas (AI) |
| [ENDPOINT_NORMALIZATION.md](ENDPOINT_NORMALIZATION.md) | Mapeo local ↔ remoto legacy |
| [ESTRATEGIA_DEPLOY_RAMAS.md](ESTRATEGIA_DEPLOY_RAMAS.md) | Estrategia conceptual de ramas y deploy |
| [CONTEXTO_AGENTE.md](CONTEXTO_AGENTE.md) | Contexto para agentes AI |

## Infraestructura (estado actual)

| Ambiente | Proyecto | Cloud SQL | DB |
|----------|----------|-----------|-----|
| DEV | new-ponti-dev | new-ponti-db-dev | new_ponti_db_dev |
| STG | new-ponti-stg | new-ponti-db-dev (cross-project) | new_ponti_db_staging |
| PROD | new-ponti-prod | new-ponti-prod-db | ponti_api_db |

## Migraciones

- **Directorio:** `migrations_v4/`
- **Make:** `make migrate-up`, `make migrate-create NAME=nombre`
