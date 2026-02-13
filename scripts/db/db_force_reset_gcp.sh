#!/usr/bin/env bash
# Fuerza reconstrucción de la base en GCP (Cloud SQL DEV):
#   - Borra el schema public y lo recrea (no requiere CREATEDB).
#   - Aplica todas las migraciones desde cero.
#
# Uso:
#   source scripts/db/db_force_reset_gcp.env && ./scripts/db/db_force_reset_gcp.sh
#   o: make db-force-reset-gcp  (después de cargar db_force_reset_gcp.env)
#
# Requiere: SRC_USER, SRC_PASS, SRC_HOST, SRC_PORT, SRC_DB, SRC_SSL
# Opcional: USE_CLOUDSQL_PROXY=1 y SRC_INSTANCE_* para usar Cloud SQL Proxy
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CREDS_FILE="${SCRIPT_DIR}/db_force_reset_gcp.env"
MIGRATIONS_PATH="${ROOT_DIR}/migrations_v4"

# Cargar credenciales GCP (origen = remoto)
if [[ -f "${CREDS_FILE}" ]]; then
  set -a
  source "${CREDS_FILE}"
  set +a
fi

# Aceptar también REMOTE_* por si se prefiere
SRC_USER="${SRC_USER:-${REMOTE_USER:-}}"
SRC_PASS="${SRC_PASS:-${REMOTE_PASSWORD:-}}"
SRC_HOST="${SRC_HOST:-${REMOTE_HOST:-}}"
SRC_PORT="${SRC_PORT:-${REMOTE_PORT:-5432}}"
SRC_DB="${SRC_DB:-${REMOTE_DB:-}}"
SRC_SSL="${SRC_SSL:-${REMOTE_SSL:-disable}}"
USE_CLOUDSQL_PROXY="${USE_CLOUDSQL_PROXY:-0}"
SRC_PROXY_PORT="${SRC_PROXY_PORT:-55432}"
SRC_INSTANCE_PROJECT="${SRC_INSTANCE_PROJECT:-new-ponti-dev}"
SRC_INSTANCE_REGION="${SRC_INSTANCE_REGION:-us-central1}"
SRC_INSTANCE_NAME="${SRC_INSTANCE_NAME:-new-ponti-db-dev}"
PROXY_CONTAINER_NAME="${PROXY_CONTAINER_NAME:-ponti-cloudsql-proxy}"

if [[ -z "${SRC_USER}" || -z "${SRC_PASS}" || -z "${SRC_HOST}" || -z "${SRC_DB}" ]]; then
  echo "Faltan credenciales. Definí SRC_* o cargá scripts/db/db_force_reset_gcp.env"
  echo "Ejemplo: source scripts/db/db_force_reset_gcp.env && ./scripts/db/db_force_reset_gcp.sh"
  exit 1
fi

log() { echo "[INFO] $*"; }
err() { echo "[ERROR] $*" >&2; }

cleanup_proxy() {
  if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${PROXY_CONTAINER_NAME}$"; then
    docker rm -f "${PROXY_CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi
}

start_proxy() {
  local conn="${SRC_INSTANCE_PROJECT}:${SRC_INSTANCE_REGION}:${SRC_INSTANCE_NAME}"
  log "Iniciando Cloud SQL Proxy (${conn}) en 127.0.0.1:${SRC_PROXY_PORT}..."
  cleanup_proxy
  docker run -d --name "${PROXY_CONTAINER_NAME}" \
    -p "127.0.0.1:${SRC_PROXY_PORT}:5432" \
    -v "${HOME}/.config/gcloud:/config" \
    gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.11.0 \
    --address 0.0.0.0 --port 5432 --gcloud-auth "${conn}" >/dev/null
  for i in {1..15}; do
    if PGPASSWORD="$SRC_PASS" pg_isready -h 127.0.0.1 -p "${SRC_PROXY_PORT}" -U "${SRC_USER}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  err "Cloud SQL Proxy no respondió."
  return 1
}

# Conexión directa o vía proxy
if [[ "${USE_CLOUDSQL_PROXY}" == "1" ]]; then
  start_proxy
  trap cleanup_proxy EXIT
  SRC_HOST="127.0.0.1"
  SRC_PORT="${SRC_PROXY_PORT}"
  SRC_SSL="disable"
fi

log "Conectando a GCP: ${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}"

# Usar Docker para psql si no está en PATH (misma red que el host para alcanzar SRC_HOST)
run_psql() {
  if command -v psql >/dev/null 2>&1; then
    PGPASSWORD="$SRC_PASS" psql -h "$SRC_HOST" -p "$SRC_PORT" -U "$SRC_USER" "$@"
  else
    docker run --rm --network host -e PGPASSWORD="$SRC_PASS" postgres:16 \
      psql -h "$SRC_HOST" -p "$SRC_PORT" -U "$SRC_USER" "$@"
  fi
}

for i in {1..20}; do
  if run_psql -d "$SRC_DB" -c "\conninfo" >/dev/null 2>&1; then
    break
  fi
  [[ $i -eq 20 ]] && { err "No se pudo conectar a la DB."; exit 1; }
  sleep 1
done

# 1) Borrar schema public y recrearlo (no requiere CREATEDB)
log "Terminando conexiones activas (si el usuario tiene permiso)..."
run_psql -d "$SRC_DB" -v ON_ERROR_STOP=0 -c "
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = current_database() AND pid <> pg_backend_pid();
" || true
log "Forzando reset: DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
run_psql -d "$SRC_DB" -v ON_ERROR_STOP=1 -c "
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
"

# 2) Aplicar migraciones (sin 000000_baseline)
MIGRATIONS_RUN_PATH="$(mktemp -d)"
cp "${MIGRATIONS_PATH}"/*.sql "${MIGRATIONS_RUN_PATH}/"
rm -f "${MIGRATIONS_RUN_PATH}/000000_baseline_schema.up.sql" \
      "${MIGRATIONS_RUN_PATH}/000000_baseline_schema.down.sql"

log "Aplicando migraciones v4 en GCP..."
docker run --rm --network host \
  -v "${MIGRATIONS_RUN_PATH}:/migrations:ro" \
  migrate/migrate:v4.17.1 \
  -path /migrations \
  -database "postgres://${SRC_USER}:${SRC_PASS}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}" \
  up

rm -rf "${MIGRATIONS_RUN_PATH}"
log "Listo: GCP DB reseteada y migraciones aplicadas."
echo "Podés cargar datos locales con: pg_dump (local) | pg_restore a esta instancia."
echo "Ejemplo data-only: pg_restore -h $SRC_HOST -p $SRC_PORT -U $SRC_USER -d $SRC_DB --data-only -v tu_dump.dump"
