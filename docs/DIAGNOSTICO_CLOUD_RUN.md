# Diagnóstico: Contenedor no inicia en Cloud Run

## 🔍 Problema

El contenedor falla al iniciar y no escucha en el puerto 8080. Esto generalmente significa que la aplicación está crasheando **antes** de iniciar el servidor HTTP.

## 📋 Pasos de Diagnóstico

### 1. Ver los logs del contenedor

**Opción A: Desde la URL del error**
- Click en la URL de logs que aparece en el error del workflow
- Busca errores al inicio del contenedor

**Opción B: Desde la terminal**
```bash
gcloud run services logs read ponti-backend-dev \
  --project=new-ponti-dev \
  --region=us-central1 \
  --limit=100
```

### 2. Verificar variables de entorno configuradas

```bash
gcloud run services describe ponti-backend-dev \
  --project=new-ponti-dev \
  --region=us-central1 \
  --format="value(spec.template.spec.containers[0].env)"
```

O usar el script de diagnóstico:
```bash
./scripts/diagnose-cloud-run.sh new-ponti-dev ponti-backend-dev us-central1
```

### 3. Variables críticas que DEBEN estar configuradas

| Variable | Valor esperado | ¿Por qué es crítica? |
|----------|----------------|---------------------|
| `GO_ENVIRONMENT` | `production` | Sin esto, la app intenta cargar `.env` y falla |
| `HTTP_SERVER_PORT` | `8080` | Puerto donde debe escuchar el servidor |
| `DB_NAME` | `branch_<slug>` o `ponti_api_db` | DB usada por el deploy |
| `DB_HOST` | IP o socket de Cloud SQL | Conexión a DB |
| `DB_USER` | Usuario de DB | Conexión a DB |
| `DB_PASSWORD` | Password de DB | Conexión a DB |
| `DB_NAME` | `ponti_api_db` | Nombre de la DB |
| `DB_PORT` | `5432` | Puerto de DB |
| `DB_SSL_MODE` | `disable` o `require` | Modo SSL |
| `DEPLOY_ENV` | `dev` | Ambiente de deploy |
| `DEPLOY_PLATFORM` | `gcp` | Plataforma |
| `X_API_KEY` | API key | Autenticación |

## 🚨 Errores Comunes

### Error: "no se pudo cargar el archivo .env base"
**Causa:** `GO_ENVIRONMENT` no está configurado  
**Solución:** Agregar `GO_ENVIRONMENT=production` en Cloud Run

### Error: "connection timed out" o "connection refused"
**Causa:** Variables de DB incorrectas o Cloud SQL no accesible  
**Solución:** Verificar `DB_HOST`, `DB_USER`, `DB_PASSWORD`, y conexión de Cloud Run a Cloud SQL

### Error: "failed to initialize schema" o "migration failed"
**Causa:** Error en migraciones o permisos de DB  
**Solución:** Ver logs para ver el error específico de migración

### Error: "router port is not Configured"
**Causa:** `HTTP_SERVER_PORT` no está configurado  
**Solución:** Ya está agregado en el workflow, pero verificar que se aplique

## 🔧 Solución Rápida

Si faltan variables críticas, configúralas manualmente:

```bash
gcloud run services update ponti-backend-dev \
  --project=new-ponti-dev \
  --region=us-central1 \
  --update-env-vars="GO_ENVIRONMENT=production,HTTP_SERVER_PORT=8080,DEPLOY_ENV=dev,DEPLOY_PLATFORM=gcp"
```

**Nota:** Esto solo agrega/actualiza estas variables, las demás se preservan.

## 📝 Próximos Pasos

1. **Ver los logs** usando la URL del error o el comando `gcloud`
2. **Identificar el error específico** (busca "Error", "Fatal", "panic")
3. **Verificar variables** usando el script de diagnóstico
4. **Corregir** según el error encontrado

## 🔗 Recursos

- [Cloud Run Troubleshooting](https://cloud.google.com/run/docs/troubleshooting)
- [Logs Viewer](https://console.cloud.google.com/logs/viewer?project=new-ponti-dev)
