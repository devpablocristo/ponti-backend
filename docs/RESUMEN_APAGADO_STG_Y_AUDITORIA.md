# Resumen: Apagado instancia STG y auditoría de costos

**Fecha:** 2026-01-30

---

## Acciones ejecutadas

| Acción | Estado |
|--------|--------|
| Migrar ponti-auth STG a instancia dev | ✅ |
| Eliminar instancia new-ponti-db-stg | ✅ |

---

## Estado final (como pedido)

| Recurso | Cantidad | Detalle |
|---------|----------|---------|
| **Cloud Run** | 6 servicios | ponti-auth, ponti-backend, ponti-frontend en DEV y STG |
| **Cloud SQL** | 1 instancia | new-ponti-db-dev (en new-ponti-dev) |
| **Databases** | 2 + postgres | new_ponti_db_dev (dev), new_ponti_db_staging (stg), postgres (sistema) |

---

## Cambios en servicios STG

### ponti-backend
- **Antes:** new-ponti-db-stg, ponti_api_db (stg), soalen-db-v3
- **Después:** new-ponti-db-dev, new_ponti_db_staging, app_stg

### ponti-auth
- **Antes:** new-ponti-stg:us-central1:new-ponti-db-stg, ponti_api_db (stg), soalen-db-v3
- **Después:** new-ponti-dev:us-central1:new-ponti-db-dev, new_ponti_db_staging, app_stg

> Al rotar la contraseña de `app_stg`, actualizar en **ambos** servicios (ponti-backend y ponti-auth).

---

## Auditoría de posibles gastos innecesarios

### ✅ Sin problemas detectados

- **Cloud SQL:** 1 instancia (dev). STG eliminada.
- **Cloud Run:** Solo servicios activos (auth, backend, frontend x2 proyectos). Sin previews huérfanos.
- **Compute Engine:** API deshabilitada en dev – no hay VMs.

### ⚠️ Revisar (bajo impacto)

| Recurso | Ubicación | Observación |
|---------|-----------|-------------|
| **Artifact Registry** | dev + stg | 6 repos por proyecto (auth-api, ponti-api, ponti-auth, ponti-backend, ponti-frontend, ui). Imágenes viejas pueden acumular. Limpiar con `gcloud artifacts docker images list` y borrar tags no usados. |
| **Cloud Storage** | backup-ponti-dev, golden-ponti-stg, run-sources-* | Buckets de backup y golden snapshot. Revisar políticas de lifecycle para borrar objetos viejos. |
| **Cloud Run revisions** | Todas las regiones | Las revisiones inactivas no cobran; se pueden borrar para orden. |

### Comandos útiles para limpieza (opcional)

```bash
# Listar imágenes viejas en un repo (ejemplo)
gcloud artifacts docker images list us-central1-docker.pkg.dev/new-ponti-dev/ponti-backend-registry/ponti-backend --include-tags

# Bucket lifecycle: configurar expiración para objetos viejos (ejemplo 90 días)
# gcloud storage buckets update gs://backup-ponti-dev --lifecycle-file=lifecycle.json
```

---

## Verificación post-apagado

- ponti-backend STG /ping: **200**
- ponti-auth STG: responde (404 en / es esperado si no hay ruta root)
- Instancia new-ponti-db-stg: **eliminada**
