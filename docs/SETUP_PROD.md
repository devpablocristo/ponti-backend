# Setup del Proyecto de Producción - New Ponti Prod

Este documento describe cómo crear y configurar el proyecto GCP de producción desde cero, usando como modelo el proyecto de desarrollo (`new-ponti-dev`).

## Nota de contexto

El deploy actual usa GitHub Actions con `dev` → `stg` → **promoción manual a prod** (mismo artefacto).  
Ver `docs/DEPLOY.md` y `docs/CONFIGURAR_VARIABLES_GITHUB.md`.

## Requisitos Previos

- Google Cloud CLI (`gcloud`) instalado y configurado
- Acceso a la organización de GCP con permisos para crear proyectos
- Acceso al repositorio de GitHub con permisos para configurar secrets/variables

## 1. Crear el Proyecto GCP

```bash
# Variables del proyecto prod
PROJECT_ID="new-ponti-prod"
PROJECT_NAME="New Ponti Prod"
BILLING_ACCOUNT_ID="<TU_BILLING_ACCOUNT_ID>"

# Crear el proyecto
gcloud projects create "$PROJECT_ID" \
  --name="$PROJECT_NAME"

# Vincular billing account
gcloud billing projects link "$PROJECT_ID" \
  --billing-account="$BILLING_ACCOUNT_ID"

# Configurar como proyecto activo
gcloud config set project "$PROJECT_ID"
```

## 2. Habilitar APIs Necesarias

```bash
# APIs requeridas
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  iamcredentials.googleapis.com \
  sts.googleapis.com \
  sqladmin.googleapis.com
```

## 3. Crear Artifact Registry

```bash
REGION="us-central1"
REGISTRY_NAME="ponti-backend-registry"

gcloud artifacts repositories create "$REGISTRY_NAME" \
  --repository-format=docker \
  --location="$REGION" \
  --description="Docker registry for ponti-backend"
```

## 4. Crear Service Accounts

### 4.1 Service Account para Cloud Run

```bash
SA_NAME="cloudrun-sa"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# Crear service account
gcloud iam service-accounts create "$SA_NAME" \
  --display-name="Cloud Run Service Account" \
  --description="Service account para Cloud Run en producción"

# Asignar roles necesarios
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.client"

# Si usas Cloud Storage u otros servicios, agregar roles adicionales aquí
```

### 4.2 Service Account para GitHub Actions (Workload Identity)

```bash
SA_GITHUB_NAME="github-actions"
SA_GITHUB_EMAIL="${SA_GITHUB_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# Crear service account
gcloud iam service-accounts create "$SA_GITHUB_NAME" \
  --display-name="GitHub Actions Deployer" \
  --description="Service account para GitHub Actions en producción"

# Asignar roles necesarios
gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${SA_GITHUB_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${SA_GITHUB_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${SA_GITHUB_EMAIL}" \
  --role="roles/iam.serviceAccountUser"
```

## 5. Configurar Workload Identity Federation

```bash
# Variables
POOL_ID="github-actions-pool"
PROVIDER_ID="github-actions-provider"
REPO="<ORG>/<REPO>"  # Ejemplo: "devpablocristo/ponti-backend"

# Obtener project number
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')

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
  --attribute-condition="assertion.repository=='$REPO'"

# Vincular repo con la Service Account
gcloud iam service-accounts add-iam-policy-binding "$SA_GITHUB_EMAIL" \
  --project="$PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/attribute.repository/${REPO}"

# Obtener el WIF_PROVIDER (necesario para GitHub)
echo "WIF_PROVIDER: projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/${POOL_ID}/providers/${PROVIDER_ID}"
echo "WIF_SERVICE_ACCOUNT: ${SA_GITHUB_EMAIL}"
```

## 6. Crear Cloud SQL (PostgreSQL)

```bash
# Variables
INSTANCE_NAME="ponti-prod-db"
DB_NAME="ponti_api_db"
DB_USER="ponti-prod-user"
DB_PASSWORD="<GENERAR_PASSWORD_SEGURO>"
REGION="us-central1"
TIER="db-f1-micro"  # Cambiar según necesidades (db-f1-micro es el más barato)

# Crear instancia Cloud SQL
gcloud sql instances create "$INSTANCE_NAME" \
  --database-version=POSTGRES_16 \
  --tier="$TIER" \
  --region="$REGION" \
  --root-password="$DB_PASSWORD" \
  --storage-type=SSD \
  --storage-size=10GB \
  --backup-start-time=03:00 \
  --enable-bin-log \
  --maintenance-window-day=SUN \
  --maintenance-window-hour=04

# Crear base de datos
gcloud sql databases create "$DB_NAME" \
  --instance="$INSTANCE_NAME"

# Crear usuario
gcloud sql users create "$DB_USER" \
  --instance="$INSTANCE_NAME" \
  --password="$DB_PASSWORD"

# Obtener IP pública (si es necesario)
gcloud sql instances describe "$INSTANCE_NAME" --format='value(ipAddresses[0].ipAddress)'
```

### 6.1 Configurar Conexión Privada (Recomendado para Prod)

Para producción, es recomendable usar Private IP en vez de IP pública:

```bash
# Habilitar Service Networking API
gcloud services enable servicenetworking.googleapis.com

# Reservar rango de IP privadas
gcloud compute addresses create google-managed-services-default \
  --global \
  --purpose=VPC_PEERING \
  --prefix-length=16 \
  --network=default

# Conectar VPC
gcloud services vpc-peerings connect \
  --service=servicenetworking.googleapis.com \
  --ranges=google-managed-services-default \
  --network=default

# Actualizar instancia para usar IP privada
gcloud sql instances patch "$INSTANCE_NAME" \
  --network=default \
  --no-assign-ip
```

## 7. Crear Servicio Cloud Run (Inicial)

```bash
SERVICE_NAME="ponti-backend-prod"
REGION="us-central1"
IMAGE_URI="$REGION-docker.pkg.dev/$PROJECT_ID/$REGISTRY_NAME/ponti-backend:prod"

# Crear servicio (inicialmente sin imagen, se actualizará en el primer deploy)
gcloud run deploy "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --image="gcr.io/cloudrun/hello" \
  --service-account="$SA_EMAIL" \
  --no-allow-unauthenticated \
  --platform=managed
```

## 8. Configurar Variables de Entorno en Cloud Run

```bash
# Actualizar variables de entorno (ajustar valores según tu configuración)
gcloud run services update "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --update-env-vars="SERVICE_NAME=ponti-api,SERVICE_VERSION=1.0,SERVICE_MAX_RETRIES=5,X_API_KEY=<PROD_API_KEY>,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=$DB_USER,DB_PASSWORD=$DB_PASSWORD,DB_HOST=/cloudsql/$PROJECT_ID:$REGION:$INSTANCE_NAME,DB_NAME=$DB_NAME,DB_SSL_MODE=require,DB_PORT=5432,MIGRATIONS_DIR=file://migrations_v4,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

### 8.1 Conectar Cloud Run a Cloud SQL

```bash
# Conectar servicio a Cloud SQL
gcloud run services update "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --add-cloudsql-instances="$PROJECT_ID:$REGION:$INSTANCE_NAME"
```

## 9. Configurar Variables en GitHub Actions

En **Settings → Secrets and variables → Actions → Variables**, agregar:

### Variables Generales (compartidas)
- `GCP_REGION`: `us-central1`
- `ARTIFACT_REGISTRY`: `ponti-backend-registry`
- `IMAGE_NAME`: `ponti-backend`
- `DEPLOY_ENV_DEV`: `dev`
- `DEPLOY_ENV_STG`: `stg`
- `DEPLOY_ENV_PROD`: `prod`
- `IMAGE_TAG_DEV`: `dev`
- `IMAGE_TAG_STG`: `stg`
- `IMAGE_TAG_PROD`: `prod`

### Variables Específicas de DEV
- `GCP_PROJECT_ID_DEV`: `new-ponti-dev`
- `SERVICE_NAME_DEV`: `ponti-backend` (o el nombre que uses en dev)
- `CLOUD_RUN_SERVICE_ACCOUNT_DEV`: `cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com`
- `WIF_PROVIDER_DEV`: `projects/<PROJECT_NUMBER_DEV>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_DEV`: `github-actions@new-ponti-dev.iam.gserviceaccount.com`

### Variables Específicas de PROD
- `GCP_PROJECT_ID_PROD`: `new-ponti-prod`
- `SERVICE_NAME_PROD`: `ponti-backend-prod`
- `CLOUD_RUN_SERVICE_ACCOUNT_PROD`: `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com`
- `WIF_PROVIDER_PROD`: `projects/<PROJECT_NUMBER_PROD>/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_PROD`: `github-actions@new-ponti-prod.iam.gserviceaccount.com`

## 10. Configurar Environment Protection en GitHub

1. Ir a **Settings → Environments**
2. Crear environment `prod`
3. Configurar:
   - **Deployment branches**: Solo `main`
   - **Required reviewers**: Agregar 1-2 personas que deben aprobar deploys a prod
   - **Wait timer**: Opcional, agregar delay de X minutos antes del deploy

## 11. Verificar Configuración

```bash
# Verificar proyecto
gcloud config get-value project

# Verificar instancia Cloud SQL
gcloud sql instances list

# Verificar servicio Cloud Run
gcloud run services list --region="$REGION"

# Verificar Artifact Registry
gcloud artifacts repositories list --location="$REGION"

# Verificar Service Accounts
gcloud iam service-accounts list
```

## 12. Primer Deploy a Producción

Una vez configurado todo:

1. Hacer merge a `main` (o push directo si es necesario)
2. El workflow de GitHub Actions debería ejecutarse automáticamente
3. Si hay protection rules, aprobar el deploy en GitHub
4. Verificar logs en Cloud Run Console

```bash
# Ver logs del servicio
gcloud run services logs read "$SERVICE_NAME" \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --limit=50
```

## Checklist Final

- [ ] Proyecto GCP creado y billing configurado
- [ ] APIs habilitadas
- [ ] Artifact Registry creado
- [ ] Service Accounts creados con roles correctos
- [ ] Workload Identity Federation configurado
- [ ] Cloud SQL instancia creada y configurada
- [ ] Base de datos y usuario creados
- [ ] Cloud Run servicio creado
- [ ] Variables de entorno configuradas en Cloud Run
- [ ] Conexión Cloud Run → Cloud SQL configurada
- [ ] Variables de GitHub Actions configuradas
- [ ] Environment protection configurado en GitHub
- [ ] Primer deploy exitoso

## Notas Importantes

1. **Seguridad en Prod**: El servicio Cloud Run en prod **NO** debe ser público (`--no-allow-unauthenticated`)
2. **SSL**: En prod usar `DB_SSL_MODE=require` (no `disable`)
3. **Backups**: Cloud SQL tiene backups automáticos, pero verificar la configuración
4. **Monitoreo**: Configurar alertas en Cloud Monitoring para errores y latencia
5. **Costos**: Monitorear costos, especialmente Cloud SQL y Cloud Run

## Troubleshooting

### Error: "Permission denied" en GitHub Actions
- Verificar que WIF_PROVIDER y WIF_SERVICE_ACCOUNT estén correctos
- Verificar que el repo en el attribute-condition coincida

### Error: "Cloud SQL connection failed"
- Verificar que Cloud Run tenga el Cloud SQL instance conectado
- Verificar que el service account tenga `roles/cloudsql.client`
- Verificar que DB_HOST use el formato Unix socket: `/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME`

### Error: "Image not found" en deploy
- Verificar que la imagen se haya pusheado correctamente a Artifact Registry
- Verificar permisos del service account de GitHub Actions
