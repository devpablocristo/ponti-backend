# Plan: Unificar DEV/STG en una sola Cloud SQL Instance

**TL;DR**: Migrar staging a la instancia `new-ponti-db-dev` (PROJECT_DEV), creando DB `new_ponti_db_staging` y usuario `app_stg` con grants aislados. DEV usa `new_ponti_db_dev` con `soalen-db-v3`. Cross-project access para Cloud Run STG.

> **Estado (2026-01-30):** ✅ Ejecutado. DB `new_ponti_db_staging` creada, ponti-backend y ponti-auth STG migrados, instancia `new-ponti-db-stg` eliminada.

---

## Datos conocidos (desde repo)

| Variable | Valor |
|----------|-------|
| PROJECT_DEV | new-ponti-dev |
| PROJECT_STG | new-ponti-stg |
| INSTANCE | new-ponti-db-dev |
| REGION | us-central1 |
| CONNECTION_NAME | new-ponti-dev:us-central1:new-ponti-db-dev |
| Cloud Run service (ambos) | ponti-backend |
| Cloud Run SA DEV | cloudrun-sa@new-ponti-dev.iam.gserviceaccount.com |
| Cloud Run SA STG | cloudrun-sa@new-ponti-stg.iam.gserviceaccount.com |
| Instancia STG vieja | ~~new-ponti-db-stg~~ **eliminada** |

---

## Estado actual (post-ejecución)

- **DEV:** `new_ponti_db_dev` (renombrado desde ponti_api_db)
- **STG:** `new_ponti_db_staging`
- **Usuarios:** soalen-db-v3 (dev), app_stg (stg), postgres (admin)

---

# 1. PRECHECK (Checklist de preflight)

Ejecutar en orden. Si algo falla, **no continuar**.

```bash
# --- 1.1 Proyecto activo ---
gcloud config get-value project
# Debe mostrar new-ponti-dev o el proyecto donde vas a operar

# --- 1.2 Autenticación ---
gcloud auth list
# Verificar que hay cuenta activa (ACTIVE)

# --- 1.3 Permisos en PROJECT_DEV ---
gcloud projects get-iam-policy new-ponti-dev --flatten="bindings[].members" \
  --filter="bindings.members:$(gcloud config get-value account)" --format="table(bindings.role)"
# Debe incluir roles/cloudsql.admin o roles/owner o similar

# --- 1.4 APIs habilitadas (PROJECT_DEV) ---
gcloud services list --enabled --project=new-ponti-dev --filter="name:sqladmin" --format="value(name)"
# Debe mostrar sqladmin.googleapis.com

# --- 1.5 Región de la instancia ---
gcloud sql instances describe new-ponti-db-dev --project=new-ponti-dev --format="value(region)"
# Esperado: us-central1

# --- 1.6 Connection string ---
gcloud sql instances describe new-ponti-db-dev --project=new-ponti-dev --format="value(connectionName)"
# Esperado: new-ponti-dev:us-central1:new-ponti-db-dev

# --- 1.7 Permisos en PROJECT_STG (para IAM binding) ---
gcloud projects get-iam-policy new-ponti-stg --flatten="bindings[].members" \
  --filter="bindings.members:$(gcloud config get-value account)" --format="table(bindings.role)"
# Necesitás roles/resourcemanager.projectIamAdmin o roles/owner en DEV para dar cloudsql.client al SA de STG
```

---

# 2. COMMANDS (Comandos gcloud/SQL comentados)

## 2a) Obtener metadata de la instancia

```bash
# PROJECT_DEV
export PROJECT_DEV=new-ponti-dev
export INSTANCE=new-ponti-db-dev

gcloud sql instances describe $INSTANCE --project=$PROJECT_DEV --format="yaml(connectionName,region,ipAddresses)"
```

## 2b) Backup on-demand pre-migración

```bash
# PROJECT_DEV - CRÍTICO: ejecutar ANTES de cualquier cambio
gcloud sql backups create \
  --instance=$INSTANCE \
  --project=$PROJECT_DEV \
  --description="pre-unificacion-dev-stg-$(date +%Y%m%d-%H%M)"

# Verificar que se creó
gcloud sql backups list --instance=$INSTANCE --project=$PROJECT_DEV --limit=3
```

## 2c) Listar databases y crear new_ponti_db_staging si no existe

```bash
# PROJECT_DEV
gcloud sql databases list --instance=$INSTANCE --project=$PROJECT_DEV

# Si new_ponti_db_staging NO existe:
gcloud sql databases create new_ponti_db_staging --instance=$INSTANCE --project=$PROJECT_DEV
```

**Nota histórica:** La DB dev se renombró de `ponti_api_db` a `new_ponti_db_dev` (ver [RENAME_DB_DEV.md](RENAME_DB_DEV.md)).

## 2d) Crear usuario app_stg

DEV usa `soalen-db-v3` (existente). Solo se crea `app_stg` para STG:

```bash
# PROJECT_DEV
# Generar password (guardar en secret manager o 1Password):
# APP_STG_PASS=$(openssl rand -base64 24)

# Crear app_stg (si no existe)
gcloud sql users create app_stg \
  --instance=$INSTANCE \
  --project=$PROJECT_DEV \
  --password='REEMPLAZAR_CON_APP_STG_PASS'

# Listar usuarios
gcloud sql users list --instance=$INSTANCE --project=$PROJECT_DEV
```

## 2e) Grants en PostgreSQL (REQUIERE psql)

Los grants en Postgres **no** se hacen con gcloud. Hay que conectarse con un usuario con privilegios (postgres o cloudsqlsuperuser) y ejecutar SQL.

### Conectar sin exponer credenciales (Cloud SQL Auth Proxy o gcloud sql connect)

```bash
# Opción 1: gcloud sql connect (usa IAM, no necesita password)
gcloud sql connect $INSTANCE --user=postgres --database=postgres --project=$PROJECT_DEV

# Opción 2: Cloud SQL Auth Proxy (si tenés postgres local)
# cloud-sql-proxy new-ponti-dev:us-central1:new-ponti-db-dev &
# psql "host=127.0.0.1 port=5432 user=postgres dbname=postgres sslmode=disable"
```

### SQL para grants (ejecutar dentro de psql)

Ver `docs/grants_app_stg.sql` para el script actual. Resumen:

```sql
-- ============================================================
-- GRANTS: app_stg SOLO new_ponti_db_staging
-- DEV usa soalen-db-v3 (no app_dev)
-- ============================================================

-- 1) Revocar CONNECT en postgres para que no pueda listar otras DBs
REVOKE CONNECT ON DATABASE postgres FROM app_stg;

-- 2) app_stg: SOLO new_ponti_db_staging
GRANT CONNECT ON DATABASE new_ponti_db_staging TO app_stg;
GRANT USAGE ON SCHEMA public TO app_stg;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_stg;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO app_stg;

-- 3) Schemas v4 (v4_core, v4_ssot, v4_calc, v4_report) - ejecutar conectado a new_ponti_db_staging:
GRANT USAGE ON SCHEMA v4_core TO app_stg;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_core TO app_stg;
GRANT USAGE ON SCHEMA v4_ssot TO app_stg;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_ssot TO app_stg;
GRANT USAGE ON SCHEMA v4_calc TO app_stg;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_calc TO app_stg;
GRANT SELECT ON ALL TABLES IN SCHEMA v4_calc TO app_stg;
GRANT USAGE ON SCHEMA v4_report TO app_stg;
GRANT SELECT ON ALL TABLES IN SCHEMA v4_report TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA v4_calc GRANT SELECT ON TABLES TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA v4_report GRANT SELECT ON TABLES TO app_stg;
```

**Importante**: Ejecutar los grants **conectado a new_ponti_db_staging** (`\c new_ponti_db_staging`). Los `GRANT ... ON DATABASE` se aplican desde cualquier conexión; los `ON SCHEMA public` y `ON ALL TABLES` deben ejecutarse **dentro de la DB**:

```sql
-- Conectado a new_ponti_db_staging
\c new_ponti_db_staging
GRANT USAGE ON SCHEMA public TO app_stg;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_stg;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_stg;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO app_stg;
-- Schemas v4
GRANT USAGE ON SCHEMA v4_core TO app_stg;
GRANT USAGE ON SCHEMA v4_ssot TO app_stg;
GRANT USAGE ON SCHEMA v4_calc TO app_stg;
GRANT USAGE ON SCHEMA v4_report TO app_stg;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_core TO app_stg;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_ssot TO app_stg;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA v4_calc TO app_stg;
GRANT SELECT ON ALL TABLES IN SCHEMA v4_calc TO app_stg;
GRANT SELECT ON ALL TABLES IN SCHEMA v4_report TO app_stg;
```

Para que **app_stg NO pueda conectarse a new_ponti_db_dev**: con solo dar `CONNECT` a `new_ponti_db_staging` y no dar CONNECT a `new_ponti_db_dev`, Postgres impide la conexión cruzada.

## 2f) Migrar datos desde instancia STG (obsoleto)

> La instancia `new-ponti-db-stg` fue eliminada. Esta sección queda como referencia histórica.

```bash
# PROJECT_STG - Exportar dump
export PROJECT_STG=new-ponti-stg
export BUCKET_STG=golden-ponti-stg-65243764597  # o un bucket en PROJECT_DEV
export STG_INSTANCE=new-ponti-db-stg
export GCS_PATH=gs://${BUCKET_STG}/migration-stg-to-dev-$(date +%Y%m%d).sql

# Crear bucket si no existe (en PROJECT_STG o PROJECT_DEV)
# gsutil mb -p $PROJECT_STG -l us-central1 gs://${BUCKET_STG} 2>/dev/null || true

gcloud sql export sql $STG_INSTANCE $GCS_PATH \
  --project=$PROJECT_STG \
  --database=ponti_api_db \
  --offload

# PROJECT_DEV - Importar a new_ponti_db_staging
gcloud sql import sql $INSTANCE $GCS_PATH \
  --project=$PROJECT_DEV \
  --database=new_ponti_db_staging
```

---

# 3. VERIFY (Comandos de verificación)

```bash
# --- Verificar databases ---
# PROJECT_DEV
gcloud sql databases list --instance=$INSTANCE --project=$PROJECT_DEV
# Debe listar new_ponti_db_dev y new_ponti_db_staging

# --- Verificar usuarios ---
gcloud sql users list --instance=$INSTANCE --project=$PROJECT_DEV
# Debe listar soalen-db-v3, app_stg

# --- Verificar grants (desde psql) ---
# Conectar como app_stg a new_ponti_db_staging:
# psql "host=/cloudsql/new-ponti-dev:us-central1:new-ponti-db-dev dbname=new_ponti_db_staging user=app_stg"
# \dt -> debe mostrar tablas
# \c new_ponti_db_dev -> debe FALLAR (permission denied)
```

---

# 4. CROSS-PROJECT ACCESS (Cloud Run STG → Cloud SQL en DEV)

## 4.1 Dar roles/cloudsql.client al SA de STG en el proyecto DEV

El Cloud Run de STG usa `cloudrun-sa@new-ponti-stg.iam.gserviceaccount.com`. Ese SA debe poder conectarse a la instancia Cloud SQL que está en **new-ponti-dev**.

```bash
# PROJECT_DEV - Ejecutar en el proyecto donde está la instancia
gcloud projects add-iam-policy-binding new-ponti-dev \
  --member="serviceAccount:cloudrun-sa@new-ponti-stg.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

**Verificación**:
```bash
gcloud projects get-iam-policy new-ponti-dev \
  --flatten="bindings[].members" \
  --filter="bindings.members:cloudrun-sa@new-ponti-stg.iam.gserviceaccount.com" \
  --format="table(bindings.role)"
# Debe mostrar roles/cloudsql.client
```

---

# 5. CONFIG CLOUD RUN (Comandos)

## 5.1 Staging (PROJECT_STG) - Apuntar a instancia DEV

```bash
# PROJECT_STG
export PROJECT_STG=new-ponti-stg
export GCP_REGION=us-central1
export SERVICE_NAME=ponti-backend
export CLOUDSQL_INSTANCE_DEV=new-ponti-dev:us-central1:new-ponti-db-dev
export DB_NAME=new_ponti_db_staging
export DB_USER=app_stg
# DB_PASSWORD desde Secret Manager o var

gcloud run services update $SERVICE_NAME \
  --project=$PROJECT_STG \
  --region=$GCP_REGION \
  --add-cloudsql-instances=$CLOUDSQL_INSTANCE_DEV \
  --set-env-vars="DB_HOST=/cloudsql/${CLOUDSQL_INSTANCE_DEV},DB_NAME=${DB_NAME},DB_USER=${DB_USER}"
```

**Importante**: Si actualmente STG tiene `--cloudsql-instances` apuntando a la instancia vieja, hay que **reemplazar** en vez de agregar:
```bash
# Reemplazar instancia (quita la vieja, pone la nueva)
gcloud run services update $SERVICE_NAME \
  --project=$PROJECT_STG \
  --region=$GCP_REGION \
  --clear-cloudsql-instances
gcloud run services update $SERVICE_NAME \
  --project=$PROJECT_STG \
  --region=$GCP_REGION \
  --add-cloudsql-instances=$CLOUDSQL_INSTANCE_DEV
```

## 5.2 Dev (PROJECT_DEV) - Usar new_ponti_db_dev y soalen-db-v3

```bash
# PROJECT_DEV
gcloud run services update ponti-backend \
  --project=new-ponti-dev \
  --region=us-central1 \
  --set-env-vars="DB_NAME=new_ponti_db_dev,DB_USER=soalen-db-v3"
# (DB_HOST ya apunta a la instancia local, no cambiar)
```

## 5.3 Guardrail en app (recomendación)

Si podés tocar código, agregar al inicio de la app:

```go
// Ejemplo en Go
if os.Getenv("DEPLOY_ENV") == "stg" && os.Getenv("DB_NAME") != "new_ponti_db_staging" {
    log.Fatal("GUARDRAIL: stg debe usar DB_NAME=new_ponti_db_staging")
}
if os.Getenv("DEPLOY_ENV") == "dev" && os.Getenv("DB_NAME") != "new_ponti_db_dev" {
    log.Fatal("GUARDRAIL: dev debe usar DB_NAME=new_ponti_db_dev")
}
```

Si no podés tocar código: documentar en runbook y verificar manualmente en cada deploy.

---

# 6. ROLLBACK (Plan de rollback)

## 6.1 Restaurar desde backup on-demand

```bash
# PROJECT_DEV - Listar backups
gcloud sql backups list --instance=new-ponti-db-dev --project=new-ponti-dev

# Restaurar (REEMPLAZA BACKUP_ID con el ID del backup pre-migración)
# ATENCIÓN: Esto restaura TODA la instancia al estado del backup. Perderás cambios posteriores.
gcloud sql backups restore BACKUP_ID \
  --backup-instance=new-ponti-db-dev \
  --backup-project=new-ponti-dev \
  --restore-instance=new-ponti-db-dev \
  --project=new-ponti-dev
```

## 6.2 Deshacer IAM binding (quitar cloudsql.client al SA de STG)

```bash
# PROJECT_DEV
gcloud projects remove-iam-policy-binding new-ponti-dev \
  --member="serviceAccount:cloudrun-sa@new-ponti-stg.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"
```

## 6.3 Revertir Cloud Run STG

> **Nota:** La instancia `new-ponti-db-stg` fue eliminada. Para revertir habría que recrear una instancia Cloud SQL STG separada.

---

# 7. COMANDOS DESTRUCTIVOS (NO EJECUTAR por defecto)

Los siguientes comandos son **destructivos**. Solo usarlos si tenés certeza y backup.

```sql
-- NO EJECUTAR - Elimina la database new_ponti_db_staging
-- DROP DATABASE new_ponti_db_staging;

-- NO EJECUTAR - Elimina usuario
-- gcloud sql users delete app_stg --instance=new-ponti-db-dev --project=new-ponti-dev
```

```bash
# NO EJECUTAR - Elimina backup
# gcloud sql backups delete BACKUP_ID --instance=new-ponti-db-dev --project=new-ponti-dev
```

---

# 8. Actualización de GitHub / repo_vars

Después de aplicar el plan, actualizar variables en GitHub (vars) y en `scripts/repo_vars.env`:

| Variable | DEV | STG |
|----------|-----|-----|
| CLOUDSQL_INSTANCE | new-ponti-dev:us-central1:new-ponti-db-dev | new-ponti-dev:us-central1:new-ponti-db-dev |
| DB_NAME | new_ponti_db_dev | new_ponti_db_staging |
| DB_USER | soalen-db-v3 | app_stg |
| DB_PASSWORD | DB_PASSWORD_DEV | DB_PASSWORD_STG |

---

## Resumen de orden de ejecución

1. PRECHECK
2. Backup on-demand
3. Crear DB new_ponti_db_staging (si no existe)
4. Crear usuario app_stg
5. Aplicar grants (psql, ver grants_app_stg.sql)
6. (Obsoleto) Migrar datos STG vieja → new_ponti_db_staging (instancia stg eliminada)
7. IAM: cloudsql.client para cloudrun-sa STG en PROJECT_DEV
8. Cloud Run: actualizar STG con nueva instancia + env vars
9. Cloud Run: actualizar DEV con new_ponti_db_dev + soalen-db-v3
10. VERIFY
11. Actualizar GitHub vars y secrets
