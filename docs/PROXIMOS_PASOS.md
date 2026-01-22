# Próximos Pasos - Configuración Final

## ✅ Lo que ya está hecho

- ✅ Proyecto GCP `new-ponti-prod` completamente configurado
- ✅ Cloud SQL instancia creada y funcionando
- ✅ Cloud Run servicio configurado con todas las variables
- ✅ X_API_KEY configurada: `981a01170e47a788ea2fbf84fbbfe57c02984b862a4ca7e0d3d2a4d76545ff59`
- ✅ Workload Identity Federation configurado
- ✅ Service Accounts con permisos correctos

## 📋 Pasos siguientes

### 1. Configurar Variables en GitHub Actions

Ir a tu repositorio en GitHub:
1. **Settings** → **Secrets and variables** → **Actions**
2. Ir a la pestaña **Variables**
3. Agregar/verificar estas variables:

#### Variables Generales (si no existen)
- `GCP_REGION` = `us-central1`
- `ARTIFACT_REGISTRY` = `ponti-backend-registry`
- `IMAGE_NAME` = `ponti-backend`
- `DEPLOY_ENV_DEV` = `dev`
- `DEPLOY_ENV_STG` = `stg`
- `DEPLOY_ENV_PROD` = `prod`
- `IMAGE_TAG_DEV` = `dev`
- `IMAGE_TAG_STG` = `stg`
- `IMAGE_TAG_PROD` = `prod`

#### Variables de DEV (si no existen)
- `GCP_PROJECT_ID_DEV` = `new-ponti-dev`
- `SERVICE_NAME_DEV` = `ponti-backend` (o el nombre que uses)
- `CLOUD_RUN_SERVICE_ACCOUNT_DEV` = `cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com`
- `WIF_PROVIDER_DEV` = `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_DEV` = `github-actions@new-ponti-dev.iam.gserviceaccount.com`

#### Variables de PROD (NUEVAS - agregar estas)
- `GCP_PROJECT_ID_PROD` = `new-ponti-prod`
- `SERVICE_NAME_PROD` = `ponti-backend-prod`
- `CLOUD_RUN_SERVICE_ACCOUNT_PROD` = `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com`
- `WIF_PROVIDER_PROD` = `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- `WIF_SERVICE_ACCOUNT_PROD` = `github-actions@new-ponti-prod.iam.gserviceaccount.com`

### 2. Configurar Environment Protection en GitHub

1. Ir a **Settings** → **Environments**
2. Click en **New environment**
3. Nombre: `prod`
4. Configurar:
   - **Deployment branches**: Seleccionar "Selected branches" y agregar solo `main`
   - **Required reviewers**: Agregar 1-2 personas que deben aprobar deploys a prod
   - **Wait timer**: Opcional (ej: 5 minutos de delay antes del deploy)

### 3. Verificar que el Workflow esté correcto

El workflow `.github/workflows/deploy-cloud-run.yml` ya está actualizado para usar proyectos diferentes según la rama.

### 4. Hacer Primer Deploy de Prueba

#### Opción A: Deploy manual desde GitHub Actions
1. Ir a **Actions** → **Deploy to Cloud Run**
2. Click en **Run workflow**
3. Seleccionar rama `main`
4. Click en **Run workflow**
5. Si hay protection rules, aprobar el deploy cuando aparezca

#### Opción B: Push a main (deploy automático)
```bash
# Hacer merge o push a main
git checkout main
git merge develop  # o la rama que quieras deployar
git push origin main
```

El workflow se ejecutará automáticamente y:
- Construirá la imagen Docker
- La subirá a Artifact Registry del proyecto prod
- Desplegará a Cloud Run en `new-ponti-prod`
- Si hay protection rules, requerirá aprobación

### 5. Verificar el Deploy

```bash
# Ver logs del servicio
gcloud run services logs read ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1 \
  --limit=50

# Ver detalles del servicio
gcloud run services describe ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1

# Probar endpoint (requiere autenticación)
curl -H "X-API-KEY: 981a01170e47a788ea2fbf84fbbfe57c02984b862a4ca7e0d3d2a4d76545ff59" \
     -H "X-User-Id: 123" \
     https://ponti-backend-prod-875939220111.us-central1.run.app/ping
```

## 🔍 Troubleshooting

### Si el deploy falla por permisos:
- Verificar que las variables de WIF_PROVIDER_PROD y WIF_SERVICE_ACCOUNT_PROD estén correctas
- Verificar que el repo en GitHub coincida con el configurado en WIF

### Si el deploy falla por imagen no encontrada:
- Verificar que ARTIFACT_REGISTRY esté correcto
- Verificar que la imagen se haya construido correctamente

### Si el servicio no puede conectar a Cloud SQL:
- Verificar que Cloud Run tenga el Cloud SQL instance conectado
- Verificar que DB_HOST use el formato Unix socket: `/cloudsql/new-ponti-prod:us-central1:ponti-prod-db`

## 📚 Documentación

- [SETUP_PROD.md](./SETUP_PROD.md) - Guía completa de setup
- [DEPLOY.md](./DEPLOY.md) - Guía de despliegue
- [GITHUB_SECRETS.md](./GITHUB_SECRETS.md) - Configuración de variables
- [RESUMEN_SETUP_PROD.md](./RESUMEN_SETUP_PROD.md) - Resumen con valores

## ✅ Checklist Final

- [ ] Variables de GitHub Actions configuradas (especialmente las de PROD)
- [ ] Environment `prod` creado con protection rules
- [ ] Primer deploy ejecutado (manual o automático)
- [ ] Deploy aprobado (si hay protection rules)
- [ ] Servicio funcionando correctamente
- [ ] Logs verificados sin errores
