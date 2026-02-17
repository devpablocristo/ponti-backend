# Configuración de Variables y Secrets en GitHub Actions

Esta guía refleja el **flujo actual** de deploys (DEV → STG → PROMOTE a PROD) y los workflows nuevos.

## Dónde se configuran

- **Repository variables**: valores no sensibles, comunes a todos los jobs.
- **Repository secrets**: valores sensibles (passwords, API keys).
- **Environments (`dev`, `stg`, `prod`)**: solo para reglas de protección.

## Variables de repositorio (no sensibles)

### Generales (compartidas)
| Variable | Valor |
|----------|-------|
| `ARTIFACT_REGISTRY` | `ponti-backend-registry` |
| `GCP_REGION` | `us-central1` |
| `IMAGE_NAME` | `ponti-backend` |
| `API_VERSION` | `v1` |
| `HTTP_SERVER_NAME` | `http-server` |
| `HTTP_SERVER_HOST` | `0.0.0.0` |
| `DB_TYPE` | `postgres` |
| `DB_SSL_MODE` | `disable` |
| `DB_PORT` | `5432` |
| `MIGRATIONS_DIR` | `file://migrations_v4` |
| `WORDS_SUGGESTER_LIMIT` | `100` |
| `WORDS_SUGGESTER_THRESHOLD` | `0.3` |
| `REPORT_SCHEMA` | `v4_report` |
| `SERVICE_NAME_ENV` | `ponti-api` |
| `SERVICE_VERSION_ENV` | `1.0` |
| `SERVICE_MAX_RETRIES_ENV` | `5` |

### DEV
| Variable | Valor |
|----------|-------|
| `GCP_PROJECT_ID_DEV` | `new-ponti-dev` |
| `SERVICE_NAME_DEV` | `ponti-backend` |
| `CLOUD_RUN_SERVICE_ACCOUNT_DEV` | `cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com` |
| `WIF_PROVIDER_DEV` | `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| `WIF_SERVICE_ACCOUNT_DEV` | `github-actions@new-ponti-dev.iam.gserviceaccount.com` |
| `CLOUDSQL_INSTANCE_DEV` | `new-ponti-dev:us-central1:new-ponti-db-dev` |
| `DB_NAME_DEV` | `new_ponti_db_dev` |
| `DB_USER_DEV` | `soalen-db-v3` |
| `DB_INSTANCE_NAME_DEV` | `new-ponti-db-dev` |
| `PREVIEW_SERVICE_PREFIX` | `ponti-backend-preview-` |
| `PREVIEW_BUCKET` | `backup-ponti-dev` |
| `PREVIEW_SEED_URI` | *(vacío si no hay seed fija)* |

#### Frontend (DEV)
| Variable | Descripción |
|----------|-------------|
| `CLOUD_RUN_SERVICE_FRONTEND_DEV` | Nombre del servicio Cloud Run del frontend (DEV) |
| `BASE_MANAGER_API_DEV` | URL del backend (DEV) usada por el BFF |

#### AI (DEV)
| Variable | Descripción |
|----------|-------------|
| `CLOUD_RUN_SERVICE_AI_DEV` | Nombre del servicio Cloud Run de AI (DEV) |

### STG (unificado en instancia new-ponti-db-dev, instancia vieja eliminada)
| Variable | Valor |
|----------|-------|
| `GCP_PROJECT_ID_STG` | `new-ponti-stg` |
| `SERVICE_NAME_STG` | `ponti-backend` |
| `CLOUD_RUN_SERVICE_ACCOUNT_STG` | `cloudrun-sa@new-ponti-stg.iam.gserviceaccount.com` |
| `WIF_PROVIDER_STG` | `projects/65243764597/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| `WIF_SERVICE_ACCOUNT_STG` | `github-actions@new-ponti-stg.iam.gserviceaccount.com` |
| `CLOUDSQL_INSTANCE_STG` | `new-ponti-dev:us-central1:new-ponti-db-dev` |
| `CLOUDSQL_PROJECT_STG` | `new-ponti-dev` |
| `DB_NAME_STG` | `new_ponti_db_staging` |
| `DB_USER_STG` | `app_stg` |
| `DB_INSTANCE_NAME_STG` | `new-ponti-db-dev` |
| `GOLDEN_SNAPSHOT_BUCKET` | `golden-ponti-stg-65243764597` |

#### Frontend (STG)
| Variable | Descripción |
|----------|-------------|
| `CLOUD_RUN_SERVICE_FRONTEND_STG` | Nombre del servicio Cloud Run del frontend (STG) |
| `BASE_MANAGER_API_STG` | URL del backend (STG) usada por el BFF |

#### AI (STG)
| Variable | Descripción |
|----------|-------------|
| `CLOUD_RUN_SERVICE_AI_STG` | Nombre del servicio Cloud Run de AI (STG) |

### PROD
| Variable | Valor |
|----------|-------|
| `GCP_PROJECT_ID_PROD` | `new-ponti-prod` |
| `SERVICE_NAME_PROD` | `ponti-backend` |
| `CLOUD_RUN_SERVICE_ACCOUNT_PROD` | `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com` |
| `WIF_PROVIDER_PROD` | `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| `WIF_SERVICE_ACCOUNT_PROD` | `github-actions@new-ponti-prod.iam.gserviceaccount.com` |
| `CLOUDSQL_INSTANCE_PROD` | `new-ponti-prod:us-central1:new-ponti-prod-db` |
| `DB_NAME_PROD` | `ponti_api_db` |
| `DB_USER_PROD` | `soalen-db-v3` |
| `DB_INSTANCE_NAME_PROD` | `new-ponti-prod-db` |

## Secrets de repositorio (sensibles)

| Secret | Descripción |
|--------|-------------|
| `DB_PASSWORD_DEV` | Password del usuario DB dev |
| `X_API_KEY_DEV` | API key dev |
| `DB_PASSWORD_STG` | Password de DB staging (usado por ponti-backend STG) |
| `X_API_KEY_STG` | API key stg |
| `DB_PASSWORD_PROD` | Password del usuario DB prod |
| `X_API_KEY_PROD` | API key prod |

## Environments en GitHub

Crear: `dev`, `stg`, `prod`.  
