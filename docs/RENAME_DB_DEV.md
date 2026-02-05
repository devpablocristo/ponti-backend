# Renombrar DB dev (completado)

**Fecha:** 2026-02-04  
**Cambio:** `ponti_api_db` → `new_ponti_db_dev` (coherencia con `new_ponti_db_staging`)

## Ejecutado

- [x] DB renombrada en Cloud SQL
- [x] ponti-backend y ponti-auth DEV actualizados
- [x] GitHub vars, repo_vars.env, .env, docker-compose, cmd/config

## Revertir (si hiciera falta)

```bash
# Conectar como postgres (ver GCP_DB_CREDS.md para credenciales)
PGPASSWORD='Soalen*25.' psql "host=136.112.24.122 port=5432 user=postgres dbname=postgres sslmode=disable" -c "
ALTER DATABASE new_ponti_db_dev RENAME TO ponti_api_db;
"

# Actualizar Cloud Run
gcloud run services update ponti-backend --project=new-ponti-dev --region=us-central1 --update-env-vars="DB_NAME=ponti_api_db"
gcloud run services update ponti-auth --project=new-ponti-dev --region=us-central1 --update-env-vars="DB_NAME=ponti_api_db"
```
