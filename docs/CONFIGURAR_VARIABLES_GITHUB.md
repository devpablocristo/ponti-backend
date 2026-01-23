# Configuración de Variables GitHub Actions - Guía Paso a Paso

## 📋 Resumen de Cambios

Necesitas **renombrar 5 variables** (agregar `_DEV`) y **crear 5 variables nuevas** de PROD.

---

## 🔄 PASO 1: Renombrar Variables Existentes

Ve a cada variable y **cambia el nombre** agregando `_DEV` al final. Los valores se mantienen iguales.

### Variables a Renombrar:

| Nombre Actual | Nuevo Nombre | Valor Actual (mantener) |
|---------------|--------------|-------------------------|
| `GCP_PROJECT_ID` | `GCP_PROJECT_ID_DEV` | `new-ponti-dev` |
| `SERVICE_NAME` | `SERVICE_NAME_DEV` | `ponti-backend-dev` |
| `CLOUD_RUN_SERVICE_ACCOUNT` | `CLOUD_RUN_SERVICE_ACCOUNT_DEV` | `cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com` |
| `WIF_PROVIDER` | `WIF_PROVIDER_DEV` | `projects/1087442197188/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider` |
| `WIF_SERVICE_ACCOUNT` | `WIF_SERVICE_ACCOUNT_DEV` | `github-actions@new-ponti-dev.iam.gserviceaccount.com` |

> **Nota**: El servicio en dev ahora se llama `ponti-backend-dev` (con sufijo) para mantener consistencia con `ponti-backend-prod`.

**Cómo hacerlo:**
1. Click en cada variable
2. Click en "Edit"
3. Cambiar el nombre agregando `_DEV` al final
4. Guardar
5. Eliminar la variable antigua (si existe)

---

## ➕ PASO 2: Crear Variables Nuevas de PROD

Click en **"New repository variable"** y crea estas 5 variables:

### 1. GCP_PROJECT_ID_PROD
- **Name**: `GCP_PROJECT_ID_PROD`
- **Value**: `new-ponti-prod`

### 2. SERVICE_NAME_PROD
- **Name**: `SERVICE_NAME_PROD`
- **Value**: `ponti-backend-prod`

> **Nota**: Ahora ambos servicios tienen sufijo consistente: `ponti-backend-dev` y `ponti-backend-prod`.

### 3. CLOUD_RUN_SERVICE_ACCOUNT_PROD
- **Name**: `CLOUD_RUN_SERVICE_ACCOUNT_PROD`
- **Value**: `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com`

### 4. WIF_PROVIDER_PROD
- **Name**: `WIF_PROVIDER_PROD`
- **Value**: `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`

### 5. WIF_SERVICE_ACCOUNT_PROD
- **Name**: `WIF_SERVICE_ACCOUNT_PROD`
- **Value**: `github-actions@new-ponti-prod.iam.gserviceaccount.com`

---

## ✅ Variables que NO Cambian (Se Mantienen Igual)

Estas variables ya están bien y **NO necesitan cambios**:

- ✅ `ARTIFACT_REGISTRY` = `ponti-backend-registry`
- ✅ `GCP_REGION` = `us-central1`
- ✅ `IMAGE_NAME` = `ponti-backend`
- ✅ `DEPLOY_ENV_DEV` = `dev`
- ✅ `DEPLOY_ENV_STG` = `stg`
- ✅ `DEPLOY_ENV_PROD` = `prod`
- ✅ `IMAGE_TAG_DEV` = `dev`
- ✅ `IMAGE_TAG_STG` = `stg`
- ✅ `IMAGE_TAG_PROD` = `prod`

---

## 📝 Lista Final de Variables (Para Verificar)

Después de hacer los cambios, deberías tener estas variables:

### Variables Generales (compartidas)
- `ARTIFACT_REGISTRY`
- `GCP_REGION`
- `IMAGE_NAME`
- `DEPLOY_ENV_DEV`
- `DEPLOY_ENV_STG`
- `DEPLOY_ENV_PROD`
- `IMAGE_TAG_DEV`
- `IMAGE_TAG_STG`
- `IMAGE_TAG_PROD`

### Variables de DEV
- `GCP_PROJECT_ID_DEV`
- `SERVICE_NAME_DEV`
- `CLOUD_RUN_SERVICE_ACCOUNT_DEV`
- `WIF_PROVIDER_DEV`
- `WIF_SERVICE_ACCOUNT_DEV`

### Variables de PROD
- `GCP_PROJECT_ID_PROD`
- `SERVICE_NAME_PROD`
- `CLOUD_RUN_SERVICE_ACCOUNT_PROD`
- `WIF_PROVIDER_PROD`
- `WIF_SERVICE_ACCOUNT_PROD`

**Total: 19 variables**

---

## 🔍 Verificación Rápida

Después de configurar todo, verifica que:

1. ✅ No queden variables con nombres genéricos (sin `_DEV` o `_PROD`)
2. ✅ Tienes 5 variables con sufijo `_DEV`
3. ✅ Tienes 5 variables con sufijo `_PROD`
4. ✅ Los valores de PROD apuntan a `new-ponti-prod`
5. ✅ Los valores de DEV apuntan a `new-ponti-dev`

---

## 🚨 Valores Importantes (Copiar y Pegar)

### Para WIF_PROVIDER_PROD (copiar completo):
```
projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider
```

### Para WIF_SERVICE_ACCOUNT_PROD:
```
github-actions@new-ponti-prod.iam.gserviceaccount.com
```

### Para CLOUD_RUN_SERVICE_ACCOUNT_PROD:
```
cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com
```

---

## ✅ Checklist

- [ ] Renombrar `GCP_PROJECT_ID` → `GCP_PROJECT_ID_DEV`
- [ ] Renombrar `SERVICE_NAME` → `SERVICE_NAME_DEV` (valor: `ponti-backend-dev`)

> **✅ ACTUALIZADO**: El servicio `ponti-backend-dev` ya fue creado en GCP. Solo necesitas actualizar la variable en GitHub.
- [ ] Renombrar `CLOUD_RUN_SERVICE_ACCOUNT` → `CLOUD_RUN_SERVICE_ACCOUNT_DEV`
- [ ] Renombrar `WIF_PROVIDER` → `WIF_PROVIDER_DEV`
- [ ] Renombrar `WIF_SERVICE_ACCOUNT` → `WIF_SERVICE_ACCOUNT_DEV`
- [ ] Crear `GCP_PROJECT_ID_PROD` = `new-ponti-prod`
- [ ] Crear `SERVICE_NAME_PROD` = `ponti-backend-prod`

> **⚠️ IMPORTANTE**: Si el servicio `ponti-backend` (sin sufijo) todavía existe en dev, puedes eliminarlo después de verificar que `ponti-backend-dev` funciona correctamente.
- [ ] Crear `CLOUD_RUN_SERVICE_ACCOUNT_PROD` = `cloudrun-sa@new-ponti-prod.iam.gserviceaccount.com`
- [ ] Crear `WIF_PROVIDER_PROD` = `projects/875939220111/locations/global/workloadIdentityPools/github-actions-pool/providers/github-actions-provider`
- [ ] Crear `WIF_SERVICE_ACCOUNT_PROD` = `github-actions@new-ponti-prod.iam.gserviceaccount.com`

---

## 🎯 Siguiente Paso

Una vez configuradas todas las variables:

1. Configurar **Environment Protection** en GitHub (Settings → Environments → `prod`)
2. Hacer el **primer deploy** a producción

Ver [DEPLOY.md](./DEPLOY.md) para más detalles sobre el proceso de deploy.
