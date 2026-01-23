# Configuración de GitHub Secrets para ponti-backend

## Secrets Requeridos

Configurar estos secrets en **Settings → Secrets and variables → Actions** del repositorio:

| Secret | Descripción | Valor de ejemplo |
|--------|-------------|------------------|
| *(ninguno)* | El deploy usa Workload Identity Federation | *(n/a)* |

## Workload Identity Federation (recomendado)

Se usa OIDC desde GitHub Actions, no se generan keys JSON.  
Cada proyecto (dev y prod) tiene su propio Workload Identity Pool y Provider.

### Valores en new-ponti-dev

- `WIF_PROVIDER_DEV`: `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_DEV`: `github-actions@new-ponti-dev.iam.gserviceaccount.com`

### Valores en new-ponti-prod

- `WIF_PROVIDER_PROD`: `projects/<PROJECT_NUMBER_PROD>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_PROD`: `github-actions@new-ponti-prod.iam.gserviceaccount.com`

> **Nota**: Para crear el WIF en prod, seguir los pasos en [SETUP_PROD.md](./SETUP_PROD.md)

## Variables para Actions

Configurar en **Settings → Secrets and variables → Actions**:

### Variables Generales (compartidas entre ambientes)

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `GCP_REGION` | Región de Cloud Run | `us-central1` |
| `ARTIFACT_REGISTRY` | Repositorio de Artifact Registry | `ponti-backend-registry` |
| `IMAGE_NAME` | Nombre de la imagen Docker | `ponti-backend` |
| `DEPLOY_ENV_DEV` | Nombre del ambiente dev | `dev` |
| `DEPLOY_ENV_STG` | Nombre del ambiente stg | `stg` |
| `DEPLOY_ENV_PROD` | Nombre del ambiente prod | `prod` |
| `IMAGE_TAG_DEV` | Tag de imagen para dev | `dev` |
| `IMAGE_TAG_STG` | Tag de imagen para stg | `stg` |
| `IMAGE_TAG_PROD` | Tag de imagen para prod | `prod` |

### Variables Específicas de DEV

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `GCP_PROJECT_ID_DEV` | ID del proyecto GCP de desarrollo | `new-ponti-dev` |
| `SERVICE_NAME_DEV` | Nombre del servicio en Cloud Run dev | `ponti-backend` |
| `CLOUD_RUN_SERVICE_ACCOUNT_DEV` | Service Account para Cloud Run dev | `cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com` |
| `WIF_PROVIDER_DEV` | Workload Identity Provider para dev | `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| `WIF_SERVICE_ACCOUNT_DEV` | Service Account para Workload Identity dev | `github-actions@new-ponti-dev.iam.gserviceaccount.com` |

### Variables Específicas de PROD

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `GCP_PROJECT_ID_PROD` | ID del proyecto GCP de producción | `new-ponti-prod` |
| `SERVICE_NAME_PROD` | Nombre del servicio en Cloud Run prod | `ponti-backend-prod` |
| `CLOUD_RUN_SERVICE_ACCOUNT_PROD` | Service Account para Cloud Run prod | `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com` |
| `WIF_PROVIDER_PROD` | Workload Identity Provider para prod | `projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| `WIF_SERVICE_ACCOUNT_PROD` | Service Account para Workload Identity prod | `github-actions@new-ponti-prod.iam.gserviceaccount.com` |

> **Nota**: El workflow selecciona automáticamente las variables correctas según la rama desplegada (`develop` → dev, `main` → prod).

### Crear Workload Identity Pool y Provider

```bash
# Variables
PROJECT_ID=<PROJECT_ID>
POOL_ID=<POOL_ID>
PROVIDER_ID=<PROVIDER_ID>
REPO=<ORG>/<REPO>

# Crear pool
gcloud iam workload-identity-pools create "$POOL_ID" \
  --project="$PROJECT_ID" \
  --location="global" \
  --display-name="GitHub Actions Pool"

# Crear provider
gcloud iam workload-identity-pools providers create-oidc "$PROVIDER_ID" \
  --project="$PROJECT_ID" \
  --location="global" \
  --workload-identity-pool="$POOL_ID" \
  --display-name="GitHub Actions Provider" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='<ORG>/<REPO>'"

# Vincular repo con la Service Account
gcloud iam service-accounts add-iam-policy-binding \
  "<SERVICE_ACCOUNT_EMAIL>" \
  --project="$PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')/locations/global/workloadIdentityPools/$POOL_ID/attribute.repository/$REPO"
```

> **Nota**: Esta opción evita `GCP_SA_KEY` y es compatible con la política que bloquea keys.

## Crear Service Account Key (solo si se permite)

```bash
# Crear Service Account (si no existe)
gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions Deployer" \
  --project=<PROJECT_ID>

# Asignar roles necesarios
gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:github-actions@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:github-actions@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding <PROJECT_ID> \
  --member="serviceAccount:github-actions@<PROJECT_ID>.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Generar key JSON
gcloud iam service-accounts keys create github-actions-key.json \
  --iam-account=github-actions@<PROJECT_ID>.iam.gserviceaccount.com \
  --project=<PROJECT_ID>

# El contenido de github-actions-key.json va en el secret GCP_SA_KEY
cat github-actions-key.json
```

## Ramas y Ambientes

| Rama | Proyecto GCP | Tag de imagen | DEPLOY_ENV | Servicio Cloud Run |
|------|--------------|---------------|------------|-------------------|
| `develop` | `new-ponti-dev` | `dev` | `dev` | `ponti-backend` |
| `staging` | `new-ponti-dev` | `stg` | `stg` | `ponti-backend` |
| `main` | `new-ponti-prod` | `prod` | `prod` | `ponti-backend-prod` |

## Flujo de Deploy

```
push to develop → build → push :dev  → deploy a new-ponti-dev (DEPLOY_ENV=dev)
push to main    → build → push :prod → deploy a new-ponti-prod (DEPLOY_ENV=prod) [requiere aprobación]
workflow_dispatch (manual) → preview en dev con DB por rama
```

> **Importante**: 
> - Deploys a `main` van al proyecto **`new-ponti-prod`** (aislado de dev)
> - Deploys a `main` requieren aprobación si hay environment protection configurado
> - El servicio en prod es privado (`--no-allow-unauthenticated`)

## Arquitectura de Proyectos

El sistema usa **dos proyectos GCP separados** para aislamiento completo:

- **`new-ponti-dev`**: Desarrollo y staging
  - Cloud SQL con IP pública o privada
  - Servicio público
  - Sin aprobaciones requeridas

- **`new-ponti-prod`**: Producción
  - Cloud SQL con IP privada (recomendado)
  - Servicio privado
  - Requiere aprobación para deploy

## Estrategia recomendada: preview por rama (DB por rama)

Para evitar conflictos de migraciones cuando se deployan ramas con diferentes versiones:

- `rama x` → **DB rama x** (preview en proyecto dev)
- `develop` → **DB dev** (proyecto dev)
- `main` → **DB prod** (proyecto prod)

**Nombres actuales:**
- Servicio preview: `ponti-backend-dev-preview-<branch-slug>`
- DB preview: `branch_<branch_slug>`

**Limpieza:**
- Automatica al cerrar PR (merge o close)
- Cron semanal como respaldo

## Variables de aplicación en Cloud Run

Las variables de la aplicación se configuran en el servicio de Cloud Run y **no** en GitHub Actions.

### Para DEV:
```bash
gcloud run services update ponti-backend \
  --project=new-ponti-dev \
  --region=us-central1 \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=dev,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=***,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=***,DB_PASSWORD=***,DB_HOST=***,DB_NAME=***,DB_SSL_MODE=disable,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

### Para PROD:
```bash
gcloud run services update ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1 \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=prod,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=***,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=***,DB_PASSWORD=***,DB_HOST=/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME,DB_NAME=***,DB_SSL_MODE=require,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

> **Nota**: En prod, `DB_HOST` debe usar el formato Unix socket para Cloud SQL y `DB_SSL_MODE=require`.

## Documentación Relacionada

- [DEPLOY.md](./DEPLOY.md) - Guía general de despliegue
- [SETUP_PROD.md](./SETUP_PROD.md) - Guía completa para crear y configurar el proyecto de producción
