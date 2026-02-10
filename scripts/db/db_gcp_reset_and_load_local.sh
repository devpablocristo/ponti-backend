#!/usr/bin/env bash
# Después del merge a develop:
#   1) Reset en GCP (DROP schema + migraciones)
#   2) Dump de la DB local
#   3) Restore solo datos en la DB de GCP
#
# Uso: make db-gcp-reset-and-load-local
#   o: ./scripts/db/db_gcp_reset_and_load_local.sh
#
# Requiere: .env (conexión local) y scripts/gcp-db-creds.env (conexión GCP)
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
CREDS_FILE="${ROOT_DIR}/scripts/gcp-db-creds.env"
DUMP_FILE="${ROOT_DIR}/.local_to_gcp_$(date +%Y%m%d_%H%M%S).dump"

log() { echo "[INFO] $*"; }
err() { echo "[ERROR] $*" >&2; }

# 1) Conexión local (desde .env)
if [[ ! -f "${ENV_FILE}" ]]; then
  err "Falta .env (conexión local)."
  exit 1
fi
set -a
source "${ENV_FILE}"
set +a
LOCAL_HOST="${DB_HOST:-127.0.0.1}"
LOCAL_PORT="${DB_PORT:-5432}"
LOCAL_USER="${DB_USER:-admin}"
LOCAL_PASS="${DB_PASSWORD:-admin}"
LOCAL_DB="${DB_NAME:-new_ponti_db_dev}"

# 2) Conexión GCP (desde gcp-db-creds.env)
if [[ ! -f "${CREDS_FILE}" ]]; then
  err "Falta scripts/gcp-db-creds.env (conexión GCP)."
  exit 1
fi
set -a
source "${CREDS_FILE}"
set +a
REMOTE_HOST="${SRC_HOST}"
REMOTE_PORT="${SRC_PORT:-5432}"
REMOTE_USER="${SRC_USER}"
REMOTE_PASS="${SRC_PASS}"
REMOTE_DB="${SRC_DB}"

if [[ -z "${REMOTE_HOST}" || -z "${REMOTE_USER}" || -z "${REMOTE_DB}" ]]; then
  err "En gcp-db-creds.env faltan SRC_HOST, SRC_USER o SRC_DB."
  exit 1
fi

# 3) Reset en GCP + migraciones
log "Paso 1/3: Reset y migraciones en GCP..."
"${ROOT_DIR}/scripts/db/db_force_reset_gcp.sh"

# Re-cargar credenciales GCP por si el script anterior no las exportó en este shell
source "${CREDS_FILE}"
REMOTE_HOST="${SRC_HOST}"
REMOTE_PORT="${SRC_PORT:-5432}"
REMOTE_USER="${SRC_USER}"
REMOTE_PASS="${SRC_PASS}"
REMOTE_DB="${SRC_DB}"

# 4) Dump desde local (solo datos)
log "Paso 2/3: Dump desde DB local (${LOCAL_HOST}:${LOCAL_PORT}/${LOCAL_DB})..."
if command -v pg_dump >/dev/null 2>&1; then
  PGPASSWORD="${LOCAL_PASS}" pg_dump -h "${LOCAL_HOST}" -p "${LOCAL_PORT}" -U "${LOCAL_USER}" -d "${LOCAL_DB}" \
    --no-owner --no-acl -F c --data-only -f "${DUMP_FILE}"
else
  docker run --rm --network host \
    -e PGPASSWORD="${LOCAL_PASS}" \
    -v "${ROOT_DIR}:${ROOT_DIR}" -w "${ROOT_DIR}" \
    postgres:16 \
    pg_dump -h "${LOCAL_HOST}" -p "${LOCAL_PORT}" -U "${LOCAL_USER}" -d "${LOCAL_DB}" \
    --no-owner --no-acl -F c --data-only -f "${DUMP_FILE}"
fi

if [[ ! -s "${DUMP_FILE}" ]]; then
  err "El dump está vacío. ¿La DB local tiene datos y está levantada?"
  rm -f "${DUMP_FILE}"
  exit 1
fi
log "Dump guardado: ${DUMP_FILE}"

# 5) Restore solo datos en GCP
log "Paso 3/3: Restore (solo datos) en GCP (${REMOTE_HOST}:${REMOTE_PORT}/${REMOTE_DB})..."
if command -v pg_restore >/dev/null 2>&1; then
  PGPASSWORD="${REMOTE_PASS}" pg_restore -h "${REMOTE_HOST}" -p "${REMOTE_PORT}" -U "${REMOTE_USER}" -d "${REMOTE_DB}" \
    --no-owner --no-acl --data-only -v "${DUMP_FILE}" || true
else
  docker run --rm --network host \
    -e PGPASSWORD="${REMOTE_PASS}" \
    -v "${ROOT_DIR}:${ROOT_DIR}" -w "${ROOT_DIR}" \
    postgres:16 \
    pg_restore -h "${REMOTE_HOST}" -p "${REMOTE_PORT}" -U "${REMOTE_USER}" -d "${REMOTE_DB}" \
    --no-owner --no-acl --data-only -v "${DUMP_FILE}" || true
fi

log "Listo: GCP reseteada, migraciones aplicadas y datos locales cargados."
log "Dump temporal: ${DUMP_FILE} (podés borrarlo si no lo necesitás)."
