# Configuración de GitHub Secrets para ponti-backend

## Secrets Requeridos

Configurar estos secrets en **Settings → Secrets and variables → Actions** del repositorio:

| Secret | Descripción | Valor de ejemplo |
|--------|-------------|------------------|
| `GCP_SA_KEY` | JSON de la Service Account de GCP con permisos para Cloud Run y Artifact Registry | `{"type": "service_account", ...}` |
| `DB_HOST` | IP/Host de PostgreSQL | `136.112.24.122` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_USER` | Usuario de la base de datos | `soalen-db-v3` |
| `DB_PASS` | Contraseña de la base de datos | `Soalen*25.` |
| `DB_NAME` | Nombre de la base de datos | `ponti_api_db` |
| `SSL_MODE` | Modo SSL | `disable` |
| `X_API_KEY` | API Key para autenticación | `abc123secreta` |

## Crear Service Account Key

```bash
# Crear Service Account (si no existe)
gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions Deployer" \
  --project=new-ponti-dev

# Asignar roles necesarios
gcloud projects add-iam-policy-binding new-ponti-dev \
  --member="serviceAccount:github-actions@new-ponti-dev.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding new-ponti-dev \
  --member="serviceAccount:github-actions@new-ponti-dev.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding new-ponti-dev \
  --member="serviceAccount:github-actions@new-ponti-dev.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Generar key JSON
gcloud iam service-accounts keys create github-actions-key.json \
  --iam-account=github-actions@new-ponti-dev.iam.gserviceaccount.com \
  --project=new-ponti-dev

# El contenido de github-actions-key.json va en el secret GCP_SA_KEY
cat github-actions-key.json
```

> **Nota**: Si aparece el error `constraints/iam.disableServiceAccountKeyCreation`, la organización bloquea keys.  
> En ese caso hay que usar **Workload Identity Federation** en el workflow en lugar de `GCP_SA_KEY`.

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

## Variables Hardcodeadas en el Workflow

Las siguientes variables están hardcodeadas porque no cambian entre ambientes:

- `GO_ENVIRONMENT=production`
- `DEPLOY_PLATFORM=gcp`
- `APP_NAME=ponti-api`
- `APP_VERSION=1.0`
- `APP_MAX_RETRIES=5`
- `API_VERSION=v1`
- `HTTP_SERVER_NAME=http-server`
- `HTTP_SERVER_HOST=0.0.0.0`
- `MIGRATIONS_DIR=file://migrations`
- `WORDS_SUGGESTER_LIMIT=100`
- `WORDS_SUGGESTER_THRESHOLD=0.3`
- `REPORT_SCHEMA=v4_report`
