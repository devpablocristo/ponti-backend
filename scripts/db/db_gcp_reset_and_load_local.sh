#!/usr/bin/env bash
# Después del merge a develop:
#   1) Reset en GCP (DROP schema + migraciones)
#   2) Dump de la DB local
#   3) Restore solo datos en la DB de GCP
#
# Uso: make db-gcp-reset-and-load-local
#   o: ./scripts/db/db_gcp_reset_and_load_local.sh
#
# Requiere: scripts/db/db_gcp_reset_and_load_local.env (conexión local + conexión GCP)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
CREDS_FILE="${SCRIPT_DIR}/db_gcp_reset_and_load_local.env"
DUMP_FILE="${ROOT_DIR}/.local_to_gcp_$(date +%Y%m%d_%H%M%S).dump"

log() { echo "[INFO] $*"; }
err() { echo "[ERROR] $*" >&2; }

# Carga credenciales remotas desde el .env y completa faltantes usando gcloud
# (evita guardar passwords en texto plano).
load_remote_creds() {
  set -a
  source "${CREDS_FILE}"
  set +a

  # Normalizar y defaults
  SRC_PORT="${SRC_PORT:-5432}"
  SRC_SSL="${SRC_SSL:-disable}"

  # Inferir host si no está seteado
  if [[ -z "${SRC_HOST:-}" ]]; then
    local project="${SRC_INSTANCE_PROJECT:-}"
    local instance="${SRC_INSTANCE_NAME:-}"
    if [[ -n "${project}" && -n "${instance}" ]] && command -v gcloud >/dev/null 2>&1; then
      SRC_HOST="$(gcloud sql instances describe "${instance}" --project="${project}" --format='value(ipAddresses[0].ipAddress)' 2>/dev/null || true)"
      export SRC_HOST
    fi
  fi

  # Inferir password desde Secret Manager si no está seteado
  if [[ -z "${SRC_PASS:-}" ]]; then
    local secret_project="${SRC_PASS_SECRET_PROJECT:-}"
    local secret_name="${SRC_PASS_SECRET_NAME:-}"
    if [[ -n "${secret_project}" && -n "${secret_name}" ]] && command -v gcloud >/dev/null 2>&1; then
      SRC_PASS="$(gcloud secrets versions access latest --secret="${secret_name}" --project="${secret_project}" 2>/dev/null || true)"
      export SRC_PASS
    fi
  fi

  REMOTE_HOST="${SRC_HOST:-}"
  REMOTE_PORT="${SRC_PORT:-5432}"
  REMOTE_USER="${SRC_USER:-}"
  REMOTE_PASS="${SRC_PASS:-}"
  REMOTE_DB="${SRC_DB:-}"

  if [[ -z "${REMOTE_HOST}" || -z "${REMOTE_USER}" || -z "${REMOTE_DB}" ]]; then
    err "En db_gcp_reset_and_load_local.env faltan SRC_HOST/SRC_USER/SRC_DB (o no se pudieron inferir)."
    err "Tip: definir SRC_INSTANCE_PROJECT y SRC_INSTANCE_NAME para inferir SRC_HOST con gcloud."
    exit 1
  fi
  if [[ -z "${REMOTE_PASS}" ]]; then
    err "Falta SRC_PASS (o no se pudo inferir desde Secret Manager)."
    err "Tip: definir SRC_PASS_SECRET_PROJECT y SRC_PASS_SECRET_NAME para leer el secreto con gcloud."
    exit 1
  fi

  export REMOTE_HOST REMOTE_PORT REMOTE_USER REMOTE_PASS REMOTE_DB
}

# 1) Conexión local + 2) Conexión GCP (SIEMPRE desde db_gcp_reset_and_load_local.env)
if [[ ! -f "${CREDS_FILE}" ]]; then
  err "Falta scripts/db/db_gcp_reset_and_load_local.env (conexión GCP)."
  exit 1
fi
set -a
source "${CREDS_FILE}"
set +a

LOCAL_HOST="${LOCAL_DB_HOST:-}"
LOCAL_PORT="${LOCAL_DB_PORT:-}"
LOCAL_USER="${LOCAL_DB_USER:-}"
LOCAL_PASS="${LOCAL_DB_PASSWORD:-}"
LOCAL_DB="${LOCAL_DB_NAME:-}"

if [[ -z "${LOCAL_HOST}" || -z "${LOCAL_PORT}" || -z "${LOCAL_USER}" || -z "${LOCAL_PASS}" || -z "${LOCAL_DB}" ]]; then
  err "En scripts/db/db_gcp_reset_and_load_local.env faltan variables locales."
  err "Requeridas: LOCAL_DB_HOST, LOCAL_DB_PORT, LOCAL_DB_USER, LOCAL_DB_PASSWORD, LOCAL_DB_NAME."
  exit 1
fi

load_remote_creds

if [[ "${DRY_RUN:-0}" == "1" ]]; then
  log "DRY_RUN=1 (sin reset/dump/restore)"
  log "Local: ${LOCAL_USER}@${LOCAL_HOST}:${LOCAL_PORT}/${LOCAL_DB}"
  log "Remote: ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PORT}/${REMOTE_DB}"
  exit 0
fi

# 3) Reset en GCP + migraciones
log "Paso 1/3: Reset y migraciones en GCP..."
"${ROOT_DIR}/scripts/db/db_force_reset_gcp.sh"

# Re-cargar credenciales GCP por si el script anterior no las exportó en este shell
load_remote_creds

# 4) Dump desde local (solo datos)
log "Paso 2/3: Dump desde DB local (${LOCAL_HOST}:${LOCAL_PORT}/${LOCAL_DB})..."
if command -v pg_dump >/dev/null 2>&1; then
  PGPASSWORD="${LOCAL_PASS}" pg_dump -h "${LOCAL_HOST}" -p "${LOCAL_PORT}" -U "${LOCAL_USER}" -d "${LOCAL_DB}" \
    --no-owner --no-acl -F c --data-only \
    --exclude-table=schema_migrations --exclude-table=schema_migrations_lock \
    -f "${DUMP_FILE}"
else
  docker run --rm --network host \
    -e PGPASSWORD="${LOCAL_PASS}" \
    -v "${ROOT_DIR}:${ROOT_DIR}" -w "${ROOT_DIR}" \
    postgres:16 \
    pg_dump -h "${LOCAL_HOST}" -p "${LOCAL_PORT}" -U "${LOCAL_USER}" -d "${LOCAL_DB}" \
    --no-owner --no-acl -F c --data-only \
    --exclude-table=schema_migrations --exclude-table=schema_migrations_lock \
    -f "${DUMP_FILE}"
fi

if [[ ! -s "${DUMP_FILE}" ]]; then
  err "El dump está vacío. ¿La DB local tiene datos y está levantada?"
  rm -f "${DUMP_FILE}"
  exit 1
fi
log "Dump guardado: ${DUMP_FILE}"

# Para evitar conflictos con seeds de migraciones (ej auth_roles),
# vaciamos tablas antes de restaurar data-only.
run_remote_psql() {
  if command -v psql >/dev/null 2>&1; then
    PGPASSWORD="${REMOTE_PASS}" psql -h "${REMOTE_HOST}" -p "${REMOTE_PORT}" -U "${REMOTE_USER}" -d "${REMOTE_DB}" "$@"
  else
    docker run --rm --network host -e PGPASSWORD="${REMOTE_PASS}" postgres:16 \
      psql -h "${REMOTE_HOST}" -p "${REMOTE_PORT}" -U "${REMOTE_USER}" -d "${REMOTE_DB}" "$@"
  fi
}

log "Vaciando tablas remotas antes del restore (TRUNCATE CASCADE)..."
run_remote_psql -v ON_ERROR_STOP=1 <<'SQL'
DO $$
DECLARE r record;
BEGIN
  FOR r IN
    SELECT tablename
    FROM pg_tables
    WHERE schemaname='public'
      AND tablename NOT IN ('schema_migrations','schema_migrations_lock')
  LOOP
    EXECUTE format('TRUNCATE TABLE public.%I RESTART IDENTITY CASCADE;', r.tablename);
  END LOOP;
END $$;
SQL

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
