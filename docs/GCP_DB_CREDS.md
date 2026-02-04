# GCP DB Credentials

## Origen: instancia DEV (IP pública)

**Proyecto:** new-ponti-dev  
**Instancia:** new-ponti-db-dev  
**Región:** us-central1  
**IP:** 136.112.24.122

**Cambio (2026-02-04):** DB renombrada `ponti_api_db` → `new_ponti_db_dev` (coherencia con `new_ponti_db_staging`).

**Dónde se aplicó en GCP:**
- Cloud SQL: instancia `new-ponti-db-dev` (proyecto new-ponti-dev) – `ALTER DATABASE ponti_api_db RENAME TO new_ponti_db_dev`
- Cloud Run: servicios `ponti-backend` y `ponti-auth` (proyecto new-ponti-dev, región us-central1) – env `DB_NAME=new_ponti_db_dev`

---
# App / download (soalen-db-v3)
SRC_USER=soalen-db-v3
SRC_PASS='Soalen*25.'
SRC_HOST=136.112.24.122
SRC_PORT=5432
SRC_SSL=disable
SRC_DB=new_ponti_db_dev

# Postgres (superuser, reset 2026-02-04 para rename DB)
POSTGRES_USER=postgres
POSTGRES_PASS='Soalen*25.'

---
# Auth (ponti-auth en new-ponti-dev, misma instancia)
SRC_USER=soalen-db-v3
SRC_PASS='Soalen*25.'
SRC_HOST=136.112.24.122
SRC_PORT=5432
SRC_SSL=disable
SRC_DB=new_ponti_db_dev
ADMIN_USER=soalenadmin25
ADMIN_PASS=Soalen*25.