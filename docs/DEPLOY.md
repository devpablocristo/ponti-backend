# Despliegue de ponti-backend en Google Cloud Run

## Requisitos Previos

- Google Cloud CLI (`gcloud`) instalado y configurado
- Docker instalado
- Acceso a los proyectos GCP:
  - `new-ponti-dev` (desarrollo)
  - `new-ponti-prod` (producción)
- Artifact Registry configurado en cada proyecto: `ponti-backend-registry`

> **Nota**: Para setup inicial del proyecto de producción, ver [SETUP_PROD.md](./SETUP_PROD.md)

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

#### Variables Generales (compartidas)
| Variable | Descripción |
|----------|-------------|
| `GCP_REGION` | Región de Cloud Run (ej: `us-central1`) |
| `ARTIFACT_REGISTRY` | Repositorio de Artifact Registry (ej: `ponti-backend-registry`) |
| `IMAGE_NAME` | Nombre de la imagen Docker (ej: `ponti-backend`) |
| `DEPLOY_ENV_DEV` | Nombre del ambiente dev (ej: `dev`) |
| `DEPLOY_ENV_STG` | Nombre del ambiente stg (ej: `stg`) |
| `DEPLOY_ENV_PROD` | Nombre del ambiente prod (ej: `prod`) |
| `IMAGE_TAG_DEV` | Tag de imagen para dev (ej: `dev`) |
| `IMAGE_TAG_STG` | Tag de imagen para stg (ej: `stg`) |
| `IMAGE_TAG_PROD` | Tag de imagen para prod (ej: `prod`) |

#### Variables Específicas de DEV
| Variable | Descripción |
|----------|-------------|
| `GCP_PROJECT_ID_DEV` | ID del proyecto GCP de desarrollo (ej: `new-ponti-dev`) |
| `SERVICE_NAME_DEV` | Nombre del servicio en Cloud Run dev (ej: `ponti-backend`) |
| `CLOUD_RUN_SERVICE_ACCOUNT_DEV` | Service Account para Cloud Run dev |
| `WIF_PROVIDER_DEV` | Workload Identity Provider para dev |
| `WIF_SERVICE_ACCOUNT_DEV` | Service Account para Workload Identity dev |

#### Variables Específicas de PROD
| Variable | Descripción |
|----------|-------------|
| `GCP_PROJECT_ID_PROD` | ID del proyecto GCP de producción (ej: `new-ponti-prod`) |
| `SERVICE_NAME_PROD` | Nombre del servicio en Cloud Run prod (ej: `ponti-backend-prod`) |
| `CLOUD_RUN_SERVICE_ACCOUNT_PROD` | Service Account para Cloud Run prod |
| `WIF_PROVIDER_PROD` | Workload Identity Provider para prod |
| `WIF_SERVICE_ACCOUNT_PROD` | Service Account para Workload Identity prod |

> **Nota**: El workflow selecciona automáticamente las variables correctas según la rama desplegada.

### Variables de aplicación en Cloud Run

Las variables de la aplicación se configuran en el servicio de Cloud Run y **no** en GitHub Actions.

#### Para DEV:
```bash
gcloud run services update ponti-backend \
  --project=new-ponti-dev \
  --region=us-central1 \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=dev,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=***,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=***,DB_PASSWORD=***,DB_HOST=***,DB_NAME=***,DB_SSL_MODE=disable,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

#### Para PROD:
```bash
gcloud run services update ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1 \
  --update-env-vars="GO_ENVIRONMENT=production,DEPLOY_ENV=prod,DEPLOY_PLATFORM=gcp,APP_NAME=ponti-api,APP_VERSION=1.0,APP_MAX_RETRIES=5,X_API_KEY=***,API_VERSION=v1,HTTP_SERVER_NAME=http-server,HTTP_SERVER_HOST=0.0.0.0,DB_TYPE=postgres,DB_USER=***,DB_PASSWORD=***,DB_HOST=/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME,DB_NAME=***,DB_SSL_MODE=require,DB_PORT=5432,MIGRATIONS_DIR=file://migrations,WORDS_SUGGESTER_LIMIT=100,WORDS_SUGGESTER_THRESHOLD=0.3,REPORT_SCHEMA=v4_report"
```

> **Nota**: En prod, `DB_HOST` debe usar el formato Unix socket para Cloud SQL: `/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME` y `DB_SSL_MODE=require` (no `disable`).

### Deploy automático por rama

- Push a `develop` → deploy a proyecto **`new-ponti-dev`** con `DEPLOY_ENV_DEV` y `IMAGE_TAG_DEV`
- Push a `staging` → deploy a proyecto **`new-ponti-dev`** con `DEPLOY_ENV_STG` y `IMAGE_TAG_STG` (usa dev por ahora)
- Push a `main` → deploy a proyecto **`new-ponti-prod`** con `DEPLOY_ENV_PROD` y `IMAGE_TAG_PROD`

> **Importante**: 
> - Deploys a `main` requieren aprobación si hay environment protection configurado
> - El servicio en prod **NO** es público (`--no-allow-unauthenticated`)
> - Cada proyecto tiene su propia instancia de Cloud SQL y recursos aislados

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

### Arquitectura de Proyectos

El sistema usa **dos proyectos GCP separados** para aislamiento completo:

- **`new-ponti-dev`**: Desarrollo y staging
  - Rama `develop` → deploy automático
  - Rama `staging` → deploy automático (usa mismo proyecto que dev)
  - Servicio público (`--allow-unauthenticated`)
  - Cloud SQL con IP pública o privada

- **`new-ponti-prod`**: Producción
  - Rama `main` → deploy automático (requiere aprobación)
  - Servicio privado (`--no-allow-unauthenticated`)
  - Cloud SQL con IP privada (recomendado)
  - SSL requerido (`DB_SSL_MODE=require`)

### Estrategia recomendada: preview por rama (DB por rama)

Para poder deployar una rama con más migraciones y luego volver a `develop` sin romper el esquema, se recomienda aislar la base de datos por rama:

- `rama x` → **DB rama x** (preview en proyecto dev)
- `develop` → **DB dev** (proyecto dev)
- `main` → **DB prod** (proyecto prod)

**Nombre sugerido (ejemplo):**
- Servicio: `ponti-backend-<branch_slug>`
- DB: `ponti_api_db_<branch_slug>`

**Limpieza:**
- Eliminar la DB y el servicio de la rama al cerrar/mergear.
- Opcional: TTL para previews sin actividad.

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
  --allow-unauthenticated
```

> **Nota**: Las variables de aplicación se gestionan en Cloud Run (ver sección anterior).

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

## URLs y Accesos

### Desarrollo
- **Service URL**: Ver en Cloud Console del proyecto `new-ponti-dev`
- **Logs**: [Cloud Console Dev](https://console.cloud.google.com/run/detail/us-central1/ponti-backend/logs?project=new-ponti-dev)

### Producción
- **Service URL**: Ver en Cloud Console del proyecto `new-ponti-prod`
- **Logs**: [Cloud Console Prod](https://console.cloud.google.com/run/detail/us-central1/ponti-backend-prod/logs?project=new-ponti-prod)
- **Nota**: El servicio en prod es privado y requiere autenticación

## Documentación Relacionada

- [SETUP_PROD.md](./SETUP_PROD.md) - Guía completa para crear y configurar el proyecto de producción
- [GITHUB_SECRETS.md](./GITHUB_SECRETS.md) - Configuración de variables y secrets en GitHub Actions
