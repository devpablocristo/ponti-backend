# Resumen de Configuración - New Ponti Prod

## ✅ Configuración Completada

### Proyecto GCP
- **Proyecto ID**: `new-ponti-prod`
- **Project Number**: `875939220111`
- **Región**: `us-central1`

### APIs Habilitadas
- Cloud Run API
- Artifact Registry API
- Cloud Build API
- IAM Credentials API
- STS API
- SQL Admin API
- Service Networking API
- Compute Engine API

### Artifact Registry
- **Repositorio**: `ponti-backend-registry`
- **Ubicación**: `us-central1`

### Service Accounts

#### Cloud Run Service Account
- **Email**: `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com`
- **Roles**: `roles/cloudsql.client`

#### GitHub Actions Service Account
- **Email**: `github-actions@new-ponti-prod.iam.gserviceaccount.com`
- **Roles**: 
  - `roles/run.admin`
  - `roles/artifactregistry.writer`
  - `roles/iam.serviceAccountUser`

### Workload Identity Federation
- **Pool**: `github-actions-pool`
- **Provider**: `github-actions-provider`
- **WIF_PROVIDER_PROD**: `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- **WIF_SERVICE_ACCOUNT_PROD**: `github-actions@new-ponti-prod.iam.gserviceaccount.com`

### Cloud Run
- **Servicio**: `ponti-backend-prod`
- **Región**: `us-central1`
- **URL**: `https://ponti-backend-prod-875939220111.us-central1.run.app`
- **Autenticación**: Privado (`--no-allow-unauthenticated`)

### Red Privada
- **Rango IP**: `google-managed-services-default` (reservado)
- **VPC Peering**: Configurado con Service Networking

### Cloud SQL
- **Instancia**: `ponti-prod-db`
- **Estado**: ⏳ CREÁNDOSE (puede tardar 5-10 minutos)
- **Versión**: PostgreSQL 15
- **Tier**: `db-f1-micro`
- **IP**: Privada (10.8.0.3)
- **Contraseña root**: `APzpgD9NrQG5n5VF1iKEdVRFY`

## ⏳ Pasos Pendientes

### 1. Completar Configuración de Cloud SQL

Cuando Cloud SQL esté en estado `RUNNABLE`, ejecutar:

```bash
./docs/COMPLETAR_SETUP_PROD.sh
```

Este script:
- Crea la base de datos `ponti_api_db`
- Crea el usuario `ponti-prod-user`
- Conecta Cloud Run a Cloud SQL
- Configura variables de entorno en Cloud Run

**Verificar estado:**
```bash
gcloud sql instances describe ponti-prod-db --format='value(state)'
```

### 2. Configurar Variables en GitHub Actions

Ir a **Settings → Secrets and variables → Actions → Variables** y agregar:

#### Variables Generales (si no existen)
- `GCP_REGION`: `us-central1`
- `ARTIFACT_REGISTRY`: `ponti-backend-registry`
- `IMAGE_NAME`: `ponti-backend`
- `DEPLOY_ENV_DEV`: `dev`
- `DEPLOY_ENV_STG`: `stg`
- `DEPLOY_ENV_PROD`: `prod`
- `IMAGE_TAG_DEV`: `dev`
- `IMAGE_TAG_STG`: `stg`
- `IMAGE_TAG_PROD`: `prod`

#### Variables de PROD
- `GCP_PROJECT_ID_PROD`: `new-ponti-prod`
- `SERVICE_NAME_PROD`: `ponti-backend-prod`
- `CLOUD_RUN_SERVICE_ACCOUNT_PROD`: `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com`
- `WIF_PROVIDER_PROD`: `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_PROD`: `github-actions@new-ponti-prod.iam.gserviceaccount.com`

### 3. Configurar Environment Protection en GitHub

1. Ir a **Settings → Environments**
2. Crear environment `prod`
3. Configurar:
   - **Deployment branches**: Solo `main`
   - **Required reviewers**: Agregar 1-2 personas que deben aprobar deploys a prod
   - **Wait timer**: Opcional

### 4. Actualizar Variables de Entorno en Cloud Run

Después de ejecutar `COMPLETAR_SETUP_PROD.sh`, actualizar `X_API_KEY` con un valor real:

```bash
gcloud run services update ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1 \
  --update-env-vars="X_API_KEY=<TU_API_KEY_REAL>"
```

## 📋 Valores Importantes

### Para GitHub Actions
```
GCP_PROJECT_ID_PROD=new-ponti-prod
SERVICE_NAME_PROD=ponti-backend-prod
CLOUD_RUN_SERVICE_ACCOUNT_PROD=cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com
WIF_PROVIDER_PROD=projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider
WIF_SERVICE_ACCOUNT_PROD=github-actions@new-ponti-prod.iam.gserviceaccount.com
```

### Para Cloud SQL
```
INSTANCE_NAME=ponti-prod-db
DB_NAME=ponti_api_db
DB_USER=ponti-prod-user
DB_PASSWORD=APzpgD9NrQG5n5VF1iKEdVRFY
DB_HOST=/cloudsql/new-ponti-prod:us-central1:ponti-prod-db
```

## 🔍 Comandos de Verificación

```bash
# Verificar estado de Cloud SQL
gcloud sql instances describe ponti-prod-db --format='value(state)'

# Verificar servicio Cloud Run
gcloud run services describe ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1

# Ver logs de Cloud Run
gcloud run services logs read ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1 \
  --limit=50

# Listar Service Accounts
gcloud iam service-accounts list --project=new-ponti-prod

# Verificar Workload Identity Pool
gcloud iam workload-identity-pools describe github-actions-pool \
  --location=global \
  --project=new-ponti-prod
```

## 🚀 Primer Deploy a Producción

Una vez completados todos los pasos:

1. Hacer merge a `main` (o push directo)
2. El workflow de GitHub Actions se ejecutará automáticamente
3. Si hay protection rules, aprobar el deploy en GitHub
4. Verificar logs en Cloud Run Console

## 📚 Documentación Relacionada

- [SETUP_PROD.md](./SETUP_PROD.md) - Guía completa de setup
- [DEPLOY.md](./DEPLOY.md) - Guía de despliegue
- [GITHUB_SECRETS.md](./GITHUB_SECRETS.md) - Configuración de variables
- [COMPLETAR_SETUP_PROD.sh](./COMPLETAR_SETUP_PROD.sh) - Script para completar configuración
