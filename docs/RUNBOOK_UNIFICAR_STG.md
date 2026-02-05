# Runbook: Unificación STG en instancia new-ponti-db-dev

**Estado:** ✅ Completado (2026-02-03) + Instancia vieja eliminada (2026-01-30)

- DB `new_ponti_db_staging` creada e importada
- Usuario `app_stg` creado, grants aplicados
- IAM configurado (cloudrun-sa@new-ponti-stg tiene cloudsql.client en new-ponti-dev)
- Cloud Run STG (backend + **ponti-auth**) actualizado y funcionando
- Instancia `new-ponti-db-stg` **eliminada**

---

## ⚠️ Seguridad: rotar contraseña app_stg

La contraseña de `app_stg` fue expuesta durante el setup. **Recomendado:** rotarla y actualizar Cloud Run + Secret Manager:

```bash
NEW_PASS=$(openssl rand -base64 24)
gcloud sql users set-password app_stg --instance=new-ponti-db-dev --project=new-ponti-dev --password="$NEW_PASS"
gcloud run services update ponti-backend --project=new-ponti-stg --region=us-central1 --update-env-vars="DB_PASSWORD=$NEW_PASS"
gcloud run services update ponti-auth --project=new-ponti-stg --region=us-central1 --update-env-vars="DB_PASS=$NEW_PASS"
# Guardar $NEW_PASS en 1Password/Secret Manager y GitHub secret DB_PASSWORD_STG
```

---

## 1. Aplicar grants (ya ejecutado)

Conectate como postgres y ejecutá el SQL:

```bash
cd /home/pablo/Projects/Pablo/ponti-backend
gcloud sql connect new-ponti-db-dev --user=postgres --database=postgres --project=new-ponti-dev
```

Cuando pida la contraseña de postgres, usá la que configuraste al crear la instancia (o resetearla si no la tenés).

Dentro de psql:

```sql
\i docs/grants_app_stg.sql
```

O pegá el contenido de `docs/grants_app_stg.sql` manualmente.

---

## 2. Actualizar Cloud Run STG

La contraseña de `app_stg` fue generada al crear el usuario. Si no la guardaste, resetearla:

```bash
# Generar nueva contraseña
NEW_PASS=$(openssl rand -base64 24)
echo "Nueva contraseña app_stg: $NEW_PASS"

# Aplicar
gcloud sql users set-password app_stg \
  --instance=new-ponti-db-dev \
  --project=new-ponti-dev \
  --password="$NEW_PASS"
```

Luego actualizar Cloud Run:

```bash
# Reemplazar YOUR_APP_STG_PASSWORD con la contraseña real
gcloud run services update ponti-backend \
  --project=new-ponti-stg \
  --region=us-central1 \
  --clear-cloudsql-instances

gcloud run services update ponti-backend \
  --project=new-ponti-stg \
  --region=us-central1 \
  --add-cloudsql-instances=new-ponti-dev:us-central1:new-ponti-db-dev \
  --update-env-vars="DB_HOST=/cloudsql/new-ponti-dev:us-central1:new-ponti-db-dev,DB_NAME=new_ponti_db_staging,DB_USER=app_stg,DB_PASSWORD=YOUR_APP_STG_PASSWORD"
```

**Importante:** Guardar la contraseña en Secret Manager o 1Password. Actualizar también las variables de GitHub Actions si STG usa `DB_PASSWORD_STG` o similar.

---

## 3. Verificación

```bash
# Health del backend STG
curl -s https://ponti-backend-65243764597.us-central1.run.app/ping

# Verificar que usa la DB correcta (revisar logs)
gcloud run services logs read ponti-backend --project=new-ponti-stg --region=us-central1 --limit=20
```

---

## 4. Actualizar GitHub (si aplica)

`scripts/repo_vars.env` ya actualizado. Si el workflow de deploy usa secrets:

- **DB_PASSWORD_STG**: actualizar con la contraseña de `app_stg` (o la nueva si rotaste)

---

## Resumen de lo ya ejecutado

- [x] Backup on-demand
- [x] DB `new_ponti_db_staging` creada
- [x] Datos importados desde dump (ponti-db-2026-01-18)
- [x] Usuario `app_stg` creado
- [x] IAM: `cloudrun-sa@new-ponti-stg` tiene `roles/cloudsql.client` en new-ponti-dev
- [x] Cloud Run ponti-backend STG migrado
- [x] Cloud Run ponti-auth STG migrado
- [x] Instancia `new-ponti-db-stg` eliminada
