# Configuración de GitHub Secrets para ponti-backend

## Secrets Requeridos

Configurar estos secrets en **Settings → Secrets and variables → Actions** del repositorio:

| Secret | Descripción | Valor de ejemplo |
|--------|-------------|------------------|
| *(ninguno)* | El deploy usa Workload Identity Federation | *(n/a)* |

## Workload Identity Federation (recomendado)

Se usa OIDC desde GitHub Actions, no se generan keys JSON.  
Se requieren estas **variables** en GitHub:

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `WIF_PROVIDER` | Provider de Workload Identity | `projects/<PROJECT_NUMBER>/locations/global/workloadIdentityPools/<POOL_ID>/providers/<PROVIDER_ID>` |
| `WIF_SERVICE_ACCOUNT` | Service Account para deploy | `<SERVICE_ACCOUNT_EMAIL>` |

### Valores actuales en new-ponti-dev

- `WIF_PROVIDER`: `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT`: `github-actions@new-ponti-dev.iam.gserviceaccount.com`

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

| Rama | Tag de imagen | DEPLOY_ENV |
|------|---------------|------------|
| `dev` | `dev` | `dev` |
| `staging` | `stg` | `stg` |
| `main` | `prod` | `prod` |

## Flujo de Deploy

```
push to dev     → build → push :dev  → deploy (DEPLOY_ENV=dev)
push to staging → build → push :stg  → deploy (DEPLOY_ENV=stg)
push to main    → build → push :prod → deploy (DEPLOY_ENV=prod)
```

## Variables de aplicación en Cloud Run

Las variables de la aplicación se configuran en el servicio de Cloud Run y **no** en GitHub Actions:

```bash
gcloud run services update <SERVICE_NAME> \
  --project=<PROJECT_ID> \
  --region=<REGION> \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=dev,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=***,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=***,DB_PASSWORD=***,DB_HOST=***,DB_NAME=***,DB_SSL_MODE=disable,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```
