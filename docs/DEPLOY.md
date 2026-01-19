# Despliegue de ponti-backend en Google Cloud Run

## Requisitos Previos

- Google Cloud CLI (`gcloud`) instalado y configurado
- Docker instalado
- Acceso al proyecto `new-ponti-dev` en GCP
- Artifact Registry configurado: `ponti-backend-registry`

## Variables de Entorno

### Variables de Aplicación

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `GO_ENVIRONMENT` | Entorno (`production`). **Obligatorio** para evitar cargar `.env` | `production` |
| `DEPLOY_ENV` | Ambiente de despliegue (`dev`, `stg`, `prod`) | `dev` |
| `DEPLOY_PLATFORM` | Plataforma (`local`, `gcp`, `aws`) | `gcp` |
| `APP_NAME` | Nombre de la aplicación | `ponti-api` |
| `APP_VERSION` | Versión de la aplicación | `1.0` |
| `APP_MAX_RETRIES` | Máximo de reintentos | `5` |

### Variables de API

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `X_API_KEY` | API Key para autenticación | `abc123secreta` |
| `API_VERSION` | Versión de la API | `v1` |
| `HTTP_SERVER_NAME` | Nombre del servidor HTTP | `http-server` |
| `HTTP_SERVER_HOST` | Host del servidor | `0.0.0.0` |

### Variables de Base de Datos

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `DB_TYPE` | Tipo de base de datos | `postgres` |
| `DB_HOST` | IP/Host de PostgreSQL | `136.112.24.122` |
| `DB_PORT` | Puerto de PostgreSQL | `5432` |
| `DB_USER` | Usuario de la base de datos | `soalen-db-v3` |
| `DB_PASSWORD` | Contraseña de la base de datos | `****` |
| `DB_NAME` | Nombre de la base de datos | `ponti_api_db` |
| `DB_SSL_MODE` | Modo SSL | `disable` |

### Variables Adicionales

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `MIGRATIONS_DIR` | Directorio de migraciones | `file://migrations` |
| `WORDS_SUGGESTER_LIMIT` | Límite del sugeridor de palabras | `100` |
| `WORDS_SUGGESTER_THRESHOLD` | Umbral del sugeridor | `0.3` |
| `REPORT_SCHEMA` | Schema para reportes | `v4_report` |

> **Nota**: La variable `PORT` es reservada por Cloud Run y se configura automáticamente.

## Despliegue con GitHub Actions

El despliegue automático y manual se gestiona desde GitHub Actions con el workflow:

- `.github/workflows/deploy-cloud-run.yml`

### Secrets requeridos (GitHub Actions)

Configurar en **Settings → Secrets and variables → Actions**:

| Secret | Descripción |
|--------|-------------|
| *(ninguno)* | El deploy usa Workload Identity Federation |

> **Nota**: El workflow usa `environment` según el ambiente (`dev`, `stg`, `prod`).  
> Si usas secrets por ambiente, deben estar definidos en ese environment.

### Variables requeridas (GitHub Actions)

Configurar en **Settings → Secrets and variables → Actions**:

| Variable | Descripción |
|----------|-------------|
| `GCP_PROJECT_ID` | ID del proyecto GCP |
| `GCP_REGION` | Región de Cloud Run |
| `ARTIFACT_REGISTRY` | Repositorio de Artifact Registry |
| `IMAGE_NAME` | Nombre de la imagen Docker |
| `SERVICE_NAME` | Nombre del servicio en Cloud Run |
| `CLOUD_RUN_SERVICE_ACCOUNT` | Service Account para Cloud Run |
| `WIF_PROVIDER` | Workload Identity Provider |
| `WIF_SERVICE_ACCOUNT` | Service Account para Workload Identity |
| `DEPLOY_ENV_DEV` | Nombre del ambiente dev |
| `DEPLOY_ENV_STG` | Nombre del ambiente stg |
| `DEPLOY_ENV_PROD` | Nombre del ambiente prod |
| `IMAGE_TAG_DEV` | Tag de imagen para dev |
| `IMAGE_TAG_STG` | Tag de imagen para stg |
| `IMAGE_TAG_PROD` | Tag de imagen para prod |

### Variables de aplicación en Cloud Run

Las variables de la aplicación se configuran en el servicio de Cloud Run y **no** en GitHub Actions:

```bash
gcloud run services update <SERVICE_NAME> \
  --project=<PROJECT_ID> \
  --region=<REGION> \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=dev,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=***,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=***,DB_PASSWORD=***,DB_HOST=***,DB_NAME=***,DB_SSL_MODE=disable,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

### Deploy automático por rama

- Push a `develop` → deploy con `DEPLOY_ENV_DEV` y `IMAGE_TAG_DEV`
- Push a `staging` → deploy con `DEPLOY_ENV_STG` y `IMAGE_TAG_STG`
- Push a `main` → deploy con `DEPLOY_ENV_PROD` y `IMAGE_TAG_PROD`

### Deploy manual por rama (sin merge)

Permite desplegar una rama específica para validar cambios antes de mergear.

Pasos:

1. GitHub → **Actions** → **Deploy to Cloud Run**
2. **Run workflow**
3. Completar:
   - `branch`: rama a desplegar (ej. `config/gpc`)

> **Nota**: El deploy manual usa siempre `DEPLOY_ENV_DEV`. En el futuro se puede habilitar `stg`.

### Flujo recomendado

1. Deploy manual de la rama a `dev`
2. Validación y aprobación
3. Merge a `develop`
4. Deploy automático por `develop`

## Pasos de Despliegue

### 1. Construir la imagen Docker

```bash
cd ponti-backend

docker build -t us-central1-docker.pkg.dev/new-ponti-dev/ponti-backend-registry/ponti-backend:dev .
```

> **Nota**: Si el build falla por problemas de DNS, configurar Docker con DNS públicos:
> ```bash
> echo '{"dns": ["8.8.8.8", "8.8.4.4"]}' | sudo tee /etc/docker/daemon.json
> sudo systemctl restart docker
> ```

### 2. Subir la imagen a Artifact Registry

```bash
docker push us-central1-docker.pkg.dev/new-ponti-dev/ponti-backend-registry/ponti-backend:dev
```

### 3. Desplegar en Cloud Run

```bash
gcloud run deploy ponti-backend \
  --project=new-ponti-dev \
  --region=us-central1 \
  --image=us-central1-docker.pkg.dev/new-ponti-dev/ponti-backend-registry/ponti-backend:dev \
  --service-account=cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com \
  --allow-unauthenticated \
  --set-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=dev,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=abc123secreta,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=soalen-db-v3,DB_PASSWORD=Soalen*25.,DB_HOST=136.112.24.122,DB_NAME=ponti_api_db,DB_SSL_MODE=disable,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

### 4. Verificar el despliegue

```bash
# Ver logs
gcloud run services logs read ponti-backend \
  --project=new-ponti-dev \
  --region=us-central1 \
  --limit=50

# Probar endpoint
curl https://ponti-backend-1087442197188.us-central1.run.app/ping
# Respuesta esperada: {"message":"pong"}
```

## Endpoints Disponibles

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/ping` | Health check |
| GET | `/api/v1/customers` | Listar clientes |
| GET | `/api/v1/projects` | Listar proyectos |
| GET | `/api/v1/campaigns` | Listar campañas |
| GET | `/api/v1/fields` | Listar campos |
| GET | `/api/v1/lots` | Listar lotes |
| GET | `/api/v1/supplies` | Listar insumos |
| GET | `/api/v1/workorders` | Listar órdenes de trabajo |
| GET | `/api/v1/labors` | Listar labores |
| GET | `/api/v1/stock` | Listar stock |
| GET | `/api/v1/dashboard` | Dashboard |
| GET | `/api/v1/reports` | Reportes |

> Todos los endpoints requieren el header `X-API-KEY` y `X-User-Id`.

## Actualizar Variables de Entorno

Para actualizar una o más variables sin redesplegar:

```bash
gcloud run services update ponti-backend \
  --project=new-ponti-dev \
  --region=us-central1 \
  --update-env-vars="VARIABLE=nuevo_valor"
```

## Troubleshooting

### Error: "no se pudo cargar el archivo .env base"
**Causa**: `GO_ENVIRONMENT` no está configurado.  
**Solución**: Asegurarse de que `GO_ENVIRONMENT=production` esté configurado.

### Error: "connection timed out" a la base de datos
**Causa**: Cloud Run no puede acceder a la IP de la base de datos.  
**Solución**: Configurar VPC Connector o autorizar las IPs de Cloud Run en el firewall.

### Error de DNS durante el build
**Causa**: Docker no puede resolver DNS.  
**Solución**: Configurar DNS públicos en Docker (ver sección de build).

## URLs de Producción

- **Service URL**: https://ponti-backend-1087442197188.us-central1.run.app
- **Logs**: [Cloud Console](https://console.cloud.google.com/run/detail/us-central1/ponti-backend/logs?project=new-ponti-dev)
