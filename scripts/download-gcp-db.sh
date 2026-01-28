#!/usr/bin/env bash
# backup_and_restore.sh (versión endurecida)
# - Preflight: prueba conexión, espera a PG, detecta superuser
# - Descarga dump (saltable con SKIP_DUMP=1)
# - Drop/Create DB y restore en 3 pasos
set -Eeuo pipefail

### ===== Preservar destino local antes de cargar .env =====
LOCAL_DB_USER="${DB_USER:-}"
LOCAL_DB_PASSWORD="${DB_PASSWORD:-}"
LOCAL_DB_HOST="${DB_HOST:-}"
LOCAL_DB_NAME="${DB_NAME:-}"
LOCAL_DB_PORT="${DB_PORT:-}"

### ===== No cargar archivos .env por ambiente =====

### ===== Origen (GCP DEV) =====
# Defaults apuntan a la DB de dev (override por variables de entorno)
SRC_USER="${SRC_USER:-${DB_USER:-}}"
SRC_PASS="${SRC_PASS:-${DB_PASSWORD:-}}"
SRC_HOST="${SRC_HOST:-${DB_HOST:-34.176.31.249}}"
SRC_DB="${SRC_DB:-${DB_NAME:-ponti_api_db}}"
SRC_PORT="${SRC_PORT:-${DB_PORT:-5432}}"
SRC_SSL="${SRC_SSL:-${DB_SSL_MODE:-disable}}"    # disable | require | verify-full

# Cloud SQL Proxy (fallback si no hay acceso directo)
USE_CLOUDSQL_PROXY="${USE_CLOUDSQL_PROXY:-auto}" # auto | 1 | 0
SRC_INSTANCE_PROJECT="${SRC_INSTANCE_PROJECT:-new-ponti-dev}"
SRC_INSTANCE_REGION="${SRC_INSTANCE_REGION:-us-central1}"
SRC_INSTANCE_NAME="${SRC_INSTANCE_NAME:-new-ponti-db-dev}"
SRC_INSTANCE_CONN="${SRC_INSTANCE_CONN:-}"
SRC_PROXY_PORT="${SRC_PROXY_PORT:-55432}"
PROXY_CONTAINER_NAME="${PROXY_CONTAINER_NAME:-ponti-cloudsql-proxy}"

### ===== Destino (Local) =====
DB_USER="${LOCAL_DB_USER:-admin}"
DB_PASSWORD="${LOCAL_DB_PASSWORD:-admin}"
DB_HOST="${LOCAL_DB_HOST:-127.0.0.1}"
DB_NAME="${LOCAL_DB_NAME:-ponti_api_db}"
DB_PORT="${LOCAL_DB_PORT:-5432}"

# Control
DISABLE_TRIGGERS="${DISABLE_TRIGGERS:-1}"  # 1= intentar deshabilitar triggers (requiere superuser)
SKIP_DUMP="${SKIP_DUMP:-0}"                # 1= salta el pg_dump
DUMP_FILE="${DUMP_FILE:-ponti_api_db_$(date +%F).dump}"
PGDUMP_RETRIES="${PGDUMP_RETRIES:-3}"
PGDUMP_RETRY_SLEEP="${PGDUMP_RETRY_SLEEP:-5}"

log(){ echo -e "\n[INFO] $*"; }
warn(){ echo -e "\n[WARN] $*"; }
err(){ echo -e "\n[ERROR] $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "No se encontró '$1' en PATH"; exit 1; }; }
# Filtra warnings conocidos sin ocultar errores reales
filter_pg_stderr() {
  python - <<'PY'
import sys
for line in sys.stdin:
    if "transaction_timeout" in line:
        continue
    if "No existing cluster is suitable as a default target" in line:
        continue
    if "errors ignored on restore:" in line:
        continue
    sys.stderr.write(line)
PY
}

run_pg_cmd() {
  local tmp
  tmp="$(mktemp)"
  if "$@" 2> "$tmp"; then
    filter_pg_stderr < "$tmp"
    rm -f "$tmp"
    return 0
  fi
  local code=$?
  if python - "$tmp" <<'PY'
import sys
path = sys.argv[1]
had = False
with open(path, "r", encoding="utf-8", errors="replace") as fh:
    for line in fh:
        if "transaction_timeout" in line:
            continue
        if "No existing cluster is suitable as a default target" in line:
            continue
        if "errors ignored on restore:" in line:
            continue
        had = True
        sys.stderr.write(line)
sys.exit(1 if had else 0)
PY
  then
    rm -f "$tmp"
    return 0
  fi
  rm -f "$tmp"
  return "$code"
}
# trap removido para permitir que el script continúe con errores menores

### ===== Validaciones de credenciales origen =====
if [[ -z "${SRC_USER}" || -z "${SRC_PASS}" ]]; then
  err "Faltan credenciales de origen. Definí SRC_USER y SRC_PASS."
  err "Ejemplo: SRC_USER=... SRC_PASS=... ./download-gcp-db.sh"
  exit 1
fi

### ===== Chequeo binarios =====
need psql; need pg_dump; need pg_restore; need pg_isready

cleanup_proxy() {
  if docker ps -a --format '{{.Names}}' | grep -q "^${PROXY_CONTAINER_NAME}$"; then
    docker rm -f "${PROXY_CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi
}

start_proxy() {
  if [[ -z "${SRC_INSTANCE_CONN}" ]]; then
    if command -v gcloud >/dev/null 2>&1; then
      local inferred
      inferred="$(gcloud sql instances list --project="${SRC_INSTANCE_PROJECT}" --format='value(connectionName)' 2>/dev/null | tr -d '\r' | head -n 1)"
      if [[ -n "${inferred}" ]]; then
        SRC_INSTANCE_CONN="${inferred}"
      else
        SRC_INSTANCE_CONN="${SRC_INSTANCE_PROJECT}:${SRC_INSTANCE_REGION}:${SRC_INSTANCE_NAME}"
      fi
    else
      SRC_INSTANCE_CONN="${SRC_INSTANCE_PROJECT}:${SRC_INSTANCE_REGION}:${SRC_INSTANCE_NAME}"
    fi
  fi
  log "Iniciando Cloud SQL Proxy (${SRC_INSTANCE_CONN}) en 127.0.0.1:${SRC_PROXY_PORT}..."
  cleanup_proxy
  # Usa credenciales locales (gcloud) o archivo de service account si se provee.
  # CLOUDSQL_CREDENTIALS_FILE: ruta al JSON de service account (opcional)
  local proxy_args=(--address 0.0.0.0 --port 5432)
  local volume_args=("-p" "127.0.0.1:${SRC_PROXY_PORT}:5432" "-v" "${HOME}/.config/gcloud:/config")
  if [[ -n "${CLOUDSQL_CREDENTIALS_FILE:-}" ]]; then
    volume_args+=("-v" "${CLOUDSQL_CREDENTIALS_FILE}:/creds.json:ro")
    proxy_args+=("--credentials-file" "/creds.json")
  else
    if command -v gcloud >/dev/null 2>&1; then
      local access_token
      access_token="$(gcloud auth print-access-token 2>/dev/null || true)"
      if [[ -n "${access_token}" ]]; then
        proxy_args+=("--token" "${access_token}")
      else
        proxy_args+=("--gcloud-auth")
      fi
    else
      proxy_args+=("--gcloud-auth")
    fi
  fi

  docker run -d --name "${PROXY_CONTAINER_NAME}" \
    "${volume_args[@]}" \
    gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.11.0 \
    "${proxy_args[@]}" "${SRC_INSTANCE_CONN}" >/dev/null

  # Esperar a que el proxy responda
  for i in {1..15}; do
    if PGPASSWORD="$SRC_PASS" pg_isready -h "127.0.0.1" -p "${SRC_PROXY_PORT}" -U "${SRC_USER}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  err "Cloud SQL Proxy no respondió. Verificá credenciales ADC o permisos cloudsql.client."
  return 1
}

trap cleanup_proxy EXIT

### ===== Probar conexión a origen y fallback a proxy =====
log "Chequeando acceso a origen ${SRC_HOST}:${SRC_PORT}..."
if ! PGPASSWORD="$SRC_PASS" pg_isready -h "$SRC_HOST" -p "$SRC_PORT" -U "$SRC_USER" >/dev/null 2>&1; then
  if [[ "$USE_CLOUDSQL_PROXY" == "1" || "$USE_CLOUDSQL_PROXY" == "auto" ]]; then
    start_proxy
    SRC_HOST="127.0.0.1"
    SRC_PORT="${SRC_PROXY_PORT}"
    SRC_SSL="disable"
  else
    err "No se pudo conectar al origen y USE_CLOUDSQL_PROXY=0."
    exit 1
  fi
fi

### ===== Esperar a que PG esté listo y probar login =====
log "Esperando PostgreSQL en ${DB_HOST}:${DB_PORT} como ${DB_USER} ..."
for i in {1..30}; do
  if PGPASSWORD="$DB_PASSWORD" pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" >/dev/null 2>&1; then
    break
  fi
  sleep 1
  [[ $i -eq 30 ]] && { err "PG no responde o credenciales inválidas.
Solución rápida:
  - Nativo: sudo -u postgres psql -c \"ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';\"
  - Docker: docker exec -it <contenedor> psql -U postgres -c \"ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';\"
"; exit 1; }
done

# Probar conexión real (captura error de auth)
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "\conninfo" >/dev/null

### ===== Detectar si el rol es superuser =====
IS_SUPER=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -At -d postgres -c "SELECT rolsuper::int FROM pg_roles WHERE rolname='${DB_USER}';" || echo 0)
if [[ "${IS_SUPER:-0}" != "1" ]]; then
  if [[ "$DISABLE_TRIGGERS" == "1" ]]; then
    warn "El rol '${DB_USER}' NO es superuser. Cambio automático DISABLE_TRIGGERS=0 para evitar errores."
    DISABLE_TRIGGERS=0
  fi
fi

### ===== Dump desde GCP (opcional) =====
if [[ "$SKIP_DUMP" == "1" && -f "$DUMP_FILE" ]]; then
  log "SKIP_DUMP=1 → uso dump existente: ${DUMP_FILE}"
else
  log "Generando dump desde GCP -> ${DUMP_FILE}"
  for attempt in $(seq 1 "$PGDUMP_RETRIES"); do
    if PGPASSWORD="$SRC_PASS" run_pg_cmd pg_dump \
      "postgresql://${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}" \
      -F c --no-owner --no-acl -v -f "$DUMP_FILE"; then
      break
    fi
    if [[ "$attempt" -lt "$PGDUMP_RETRIES" ]]; then
      warn "pg_dump falló (intento ${attempt}/${PGDUMP_RETRIES}). Reintento en ${PGDUMP_RETRY_SLEEP}s..."
      sleep "$PGDUMP_RETRY_SLEEP"
    else
      err "pg_dump falló luego de ${PGDUMP_RETRIES} intentos."
      exit 1
    fi
  done
fi

if [[ ! -s "$DUMP_FILE" ]]; then
  err "El dump generado está vacío o incompleto: ${DUMP_FILE}"
  exit 1
fi

log "Contenido del dump (primeras 20 líneas):"
run_pg_cmd pg_restore -l "$DUMP_FILE" | head -n 20 || true

### ===== Drop/Create database destino =====
log "Terminando conexiones activas a '${DB_NAME}' (si existieran)…"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname='${DB_NAME}' AND pid<>pg_backend_pid();
" || true

log "DROP DATABASE IF EXISTS ${DB_NAME};"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${DB_NAME};"

log "CREATE DATABASE ${DB_NAME};"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${DB_NAME};"

log "Conexión a la DB nueva:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\conninfo"

### ===== Restore en 3 pasos =====
RESTORE_COMMON=(-h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --no-owner --no-privileges -v)

log "PASO 1/3 — PRE-DATA (esquema)…"
if PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" --section=pre-data "$DUMP_FILE"; then
  log "  ✅ Pre-data completado exitosamente"
else
  warn "  ⚠️  Pre-data tuvo errores (como transaction_timeout) pero continuando..."
fi

log "PASO 2/3 — DATA (solo datos)…"
if [[ "$DISABLE_TRIGGERS" == "1" ]]; then
  log "  -> con --disable-triggers (sin --single-transaction para evitar transaction_timeout)"
  if PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" --section=data --disable-triggers "$DUMP_FILE"; then
    log "  ✅ Data completado exitosamente"
  else
    warn "  ⚠️  Data tuvo errores pero continuando..."
  fi
else
  log "  -> SIN --disable-triggers (sin --single-transaction para evitar transaction_timeout)"
  if PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" --section=data "$DUMP_FILE"; then
    log "  ✅ Data completado exitosamente"
  else
    warn "  ⚠️  Data tuvo errores pero continuando..."
  fi
fi

log "PASO 3/3 — POST-DATA (índices, FKs, constraints)…"
if PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" --section=post-data "$DUMP_FILE"; then
  log "  ✅ Post-data completado exitosamente"
else
  warn "  ⚠️  Post-data tuvo errores pero continuando..."
fi

### ===== Verificación rápida =====
log "Tablas (primeras 30 filas de \dt):"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -c "\dt public.*" | sed -n '1,30p' || true

log "✅ Restauración completada con éxito en '${DB_NAME}'"
