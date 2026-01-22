# Cómo Probar que Todo Funciona

## 🧪 Opciones de Prueba

### Opción 1: Deploy Manual desde GitHub (Recomendado)

1. Ir a **Actions** → **Deploy to Cloud Run**
2. Click en **Run workflow**
3. Seleccionar rama: `main`
4. Click **Run workflow**
5. Si hay environment protection, **aprobar** cuando aparezca
6. Ver logs del workflow para verificar que todo funciona

**Qué verificar:**
- ✅ Build de imagen exitoso
- ✅ Push a Artifact Registry exitoso
- ✅ Deploy a Cloud Run exitoso
- ✅ Sin errores en los logs

---

### Opción 2: Push a Main (Deploy Automático)

```bash
# Hacer merge o push a main
git checkout main
git merge develop  # o la rama que quieras probar
git push origin main
```

El workflow se ejecutará automáticamente y:
- Construirá la imagen
- La subirá a Artifact Registry de prod
- Desplegará a `ponti-backend-prod`
- Requerirá aprobación si hay protection rules

---

### Opción 3: Probar Servicios Directamente

#### Servicio DEV:
```bash
curl -H "X-API-KEY: abc123secreta" \
     -H "X-User-Id: 123" \
     https://ponti-backend-dev-vjhlkdxpoa-uc.a.run.app/ping
```

**Respuesta esperada:** `{"message":"pong"}`

#### Servicio PROD:
```bash
curl -H "X-API-KEY: 981a01170e47a788ea2fbf84fbbfe57c02984b862a4ca7e0d3d2a4d76545ff59" \
     -H "X-User-Id: 123" \
     https://ponti-backend-prod-875939220111.us-central1.run.app/ping
```

**Respuesta esperada:** `{"message":"pong"}`

> **Nota**: El servicio de prod es privado, pero el endpoint `/ping` debería responder si la API key es correcta.

---

### Opción 4: Ver Logs

#### Logs del Workflow en GitHub:
1. Ir a **Actions**
2. Click en el workflow ejecutado
3. Revisar cada step para ver si hay errores

#### Logs de Cloud Run:
```bash
# Logs de DEV
gcloud run services logs read ponti-backend-dev \
  --project=new-ponti-dev \
  --region=us-central1 \
  --limit=50

# Logs de PROD
gcloud run services logs read ponti-backend-prod \
  --project=new-ponti-prod \
  --region=us-central1 \
  --limit=50
```

---

### Opción 5: Verificar Estado de Servicios

```bash
# Ver servicios DEV
gcloud run services list \
  --project=new-ponti-dev \
  --region=us-central1 \
  --format="table(name,status.url,status.conditions[0].status)"

# Ver servicios PROD
gcloud run services list \
  --project=new-ponti-prod \
  --region=us-central1 \
  --format="table(name,status.url,status.conditions[0].status)"
```

---

## ✅ Checklist de Verificación

Después de hacer el deploy, verifica:

- [ ] Workflow ejecutado sin errores
- [ ] Imagen construida y pusheada correctamente
- [ ] Deploy a Cloud Run exitoso
- [ ] Servicio responde al `/ping`
- [ ] Logs sin errores críticos
- [ ] Variables de entorno correctas en Cloud Run
- [ ] Conexión a Cloud SQL funcionando (si aplica)

---

## 🚨 Troubleshooting

### Si el workflow falla:
1. Revisar logs del step que falló
2. Verificar que todas las variables estén configuradas
3. Verificar permisos de Workload Identity

### Si el servicio no responde:
1. Ver logs de Cloud Run
2. Verificar que la API key sea correcta
3. Verificar que el servicio esté en estado "Ready"

### Si hay errores de conexión a DB:
1. Verificar que Cloud SQL esté conectado al servicio
2. Verificar variables de entorno (DB_HOST, DB_USER, etc.)
3. Verificar que el service account tenga permisos
