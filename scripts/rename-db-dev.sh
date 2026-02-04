#!/bin/bash
# Renombrar ponti_api_db → new_ponti_db_dev y actualizar Cloud Run
# NOTA: Este script ya fue ejecutado. La DB se renombró y Cloud Run actualizado.
# Mantener por referencia. Si necesitás revertir, crear ponti_api_db y restaurar.
#
# Conexión: usa IP pública (136.112.24.122) cuando ADC/proxy no están disponibles.
# Requiere: PGPASSWORD para usuario postgres, o SRC_HOST/SRC_PASS de GCP_DB_CREDS para soalen-db-v3 (pero postgres es necesario para ALTER DATABASE).

set -e
INSTANCE=new-ponti-db-dev
PROJECT=new-ponti-dev
# IP pública de la instancia (fallback cuando proxy/ADC fallan)
DB_HOST="${DB_HOST:-136.112.24.122}"
DB_PORT="${DB_PORT:-5432}"

echo "1. Renombrando DB..."

export PGPASSWORD="${PGPASSWORD:-$POSTGRES_PASSWORD}"
if [[ -z "${PGPASSWORD}" ]]; then
  echo "Error: Necesitás PGPASSWORD o POSTGRES_PASSWORD para usuario postgres."
  echo "Ejemplo: PGPASSWORD='tu_pass' $0"
  echo "O: gcloud sql users set-password postgres --instance=$INSTANCE --project=$PROJECT --password='TempPass'"
  exit 1
fi

psql "host=${DB_HOST} port=${DB_PORT} user=postgres dbname=postgres sslmode=disable" -v ON_ERROR_STOP=1 -c "
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'ponti_api_db' AND pid <> pg_backend_pid();
ALTER DATABASE ponti_api_db RENAME TO new_ponti_db_dev;
" || { echo "Si la DB ya se renombró, ignorar. Continuando..."; true; }

echo "2. Actualizando ponti-backend DEV..."
gcloud run services update ponti-backend --project=$PROJECT --region=us-central1 --update-env-vars="DB_NAME=new_ponti_db_dev"

echo "3. Actualizando ponti-auth DEV..."
gcloud run services update ponti-auth --project=$PROJECT --region=us-central1 --update-env-vars="DB_NAME=new_ponti_db_dev"

echo "Listo. Verificar: curl -s \$(gcloud run services describe ponti-backend --project=$PROJECT --region=us-central1 --format='value(status.url)')/ping"
