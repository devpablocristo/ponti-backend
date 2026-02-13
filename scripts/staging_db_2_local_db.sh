#!/usr/bin/env bash
# staging_db_2_local_db.sh
# - Descarga dump desde GCP STAGING y restaura solo datos en DB local (tal cual, sin cambios)
# - Origen: new_ponti_db_staging. Tratamiento: data-only, sin schema, sin renames.
#
# Requiere: SRC_PASS para el usuario de STAGING configurado en SRC_USER/DB_USER_STG.
# Opcional: scripts/staging_db_2_local_db.env con SRC_PASS, SRC_HOST, etc.
#
# Uso: SRC_PASS='...' ./scripts/staging_db_2_local_db.sh
#   o: cp scripts/staging_db_2_local_db.env.example scripts/staging_db_2_local_db.env && ./scripts/staging_db_2_local_db.sh
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

### ===== Preservar destino local antes de cargar .env =====
LOCAL_DB_USER="${DB_USER:-}"
LOCAL_DB_PASSWORD="${DB_PASSWORD:-}"
LOCAL_DB_HOST="${DB_HOST:-}"
LOCAL_DB_NAME="${DB_NAME:-}"
LOCAL_DB_PORT="${DB_PORT:-}"

### ===== Cargar .env del backend =====
ENV_FILE="${BACKEND_DIR}/.env"
if [[ -f "$ENV_FILE" ]]; then
  set -a
  source "$ENV_FILE"
  set +a
fi

### ===== Origen: STAGING (desde staging_db_2_local_db.env) =====
if [[ -f "${SCRIPT_DIR}/staging_db_2_local_db.env" ]]; then
  set -a
  source "${SCRIPT_DIR}/staging_db_2_local_db.env"
  set +a
fi

SRC_USER="${SRC_USER:-${DB_USER_STG:-soalen-db-v3}}"
SRC_PASS="${SRC_PASS:-}"
SRC_HOST="${SRC_HOST:-136.112.24.122}"
SRC_PORT="${SRC_PORT:-5432}"
SRC_DB="${SRC_DB:-${DB_NAME_STG:-new_ponti_db_staging}}"
SRC_SSL="${SRC_SSL:-disable}"

# Cloud SQL Proxy (fallback si no hay acceso directo)
USE_CLOUDSQL_PROXY="${USE_CLOUDSQL_PROXY:-auto}" # auto | 1 | 0
SRC_INSTANCE_PROJECT="${SRC_INSTANCE_PROJECT:-${CLOUDSQL_PROJECT_STG:-new-ponti-dev}}"
SRC_INSTANCE_REGION="${SRC_INSTANCE_REGION:-us-central1}"
SRC_INSTANCE_NAME="${SRC_INSTANCE_NAME:-${DB_INSTANCE_NAME_STG:-new-ponti-db-dev}}"
SRC_INSTANCE_CONN="${SRC_INSTANCE_CONN:-}"
SRC_PROXY_PORT="${SRC_PROXY_PORT:-55433}"
PROXY_CONTAINER_NAME="${PROXY_CONTAINER_NAME:-ponti-cloudsql-proxy}"

### ===== Destino (Local) =====
DB_USER="${LOCAL_DB_USER:-admin}"
DB_PASSWORD="${LOCAL_DB_PASSWORD:-admin}"
DB_HOST="${LOCAL_DB_HOST:-127.0.0.1}"
DB_NAME="${LOCAL_DB_NAME:-new_ponti_db_dev}"
DB_PORT="${LOCAL_DB_PORT:-5432}"

# Control
DISABLE_TRIGGERS="${DISABLE_TRIGGERS:-1}"  # 1= intentar deshabilitar triggers (requiere superuser)
SKIP_DUMP="${SKIP_DUMP:-0}"                # 1= salta el pg_dump
DUMP_FILE="${DUMP_FILE:-/tmp/staging_to_local_$(date +%F_%H%M%S).dump}"
PGDUMP_RETRIES="${PGDUMP_RETRIES:-3}"
PGDUMP_RETRY_SLEEP="${PGDUMP_RETRY_SLEEP:-5}"
RESTORE_MODE="${RESTORE_MODE:-data-only}" # data-only | full (solo data-only soportado para staging)
TRUNCATE_BEFORE_RESTORE="${TRUNCATE_BEFORE_RESTORE:-1}"

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

# URL-encode password para connection string (evita PGPASSFILE/PGPASSWORD que pueden fallar)
urlencode_pass() {
  python3 -c "
import urllib.parse
import sys
print(urllib.parse.quote(sys.argv[1], safe=''))
" "$1"
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

### ===== Validaciones de credenciales origen (STAGING) =====
if [[ -z "${SRC_PASS}" ]]; then
  err "SRC_PASS es requerido para el usuario de staging (${SRC_USER})."
  err "Creá scripts/staging_db_2_local_db.env con SRC_PASS=... o pasá SRC_PASS='...' $0"
  exit 1
fi

if [[ -z "${SRC_USER}" || -z "${SRC_HOST}" || -z "${SRC_DB}" || -z "${SRC_PORT}" ]]; then
  err "Faltan credenciales de origen. Definí SRC_USER, SRC_PASS, SRC_HOST, SRC_DB, SRC_PORT."
  err "Ejemplo: SRC_PASS='...' ./scripts/staging_db_2_local_db.sh"
  exit 1
fi

if [[ -z "${SRC_SSL}" ]]; then
  SRC_SSL="disable"
fi

log "Origen efectivo: ${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB} (STAGING, sslmode=${SRC_SSL})"

### ===== Chequeo binarios =====
need psql; need pg_dump; need pg_restore; need pg_isready

cleanup_proxy() {
  if docker ps -a --format '{{.Names}}' | grep -q "^${PROXY_CONTAINER_NAME}$"; then
    docker rm -f "${PROXY_CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi
}

start_proxy() {
  # Solo reutilizar si NUESTRO contenedor proxy está corriendo (evita confundir con ponti-ai u otros)
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${PROXY_CONTAINER_NAME}$"; then
    if PGPASSWORD="$SRC_PASS" pg_isready -h "127.0.0.1" -p "${SRC_PROXY_PORT}" -U "${SRC_USER}" >/dev/null 2>&1; then
      log "Proxy ya activo (${PROXY_CONTAINER_NAME}), reutilizando"
      return 0
    fi
  fi
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
# USE_CLOUDSQL_PROXY=1: forzar proxy (útil si auth directa falla)
# USE_CLOUDSQL_PROXY=auto: usar proxy solo si pg_isready falla
if [[ "$USE_CLOUDSQL_PROXY" == "1" ]]; then
  log "USE_CLOUDSQL_PROXY=1 → usando Cloud SQL Proxy (evita auth directa)"
  start_proxy
  SRC_HOST="127.0.0.1"
  SRC_PORT="${SRC_PROXY_PORT}"
  SRC_SSL="disable"
else
  log "Chequeando acceso a origen ${SRC_HOST}:${SRC_PORT}..."
  if ! PGPASSWORD="$SRC_PASS" pg_isready -h "$SRC_HOST" -p "$SRC_PORT" -U "$SRC_USER" >/dev/null 2>&1; then
    if [[ "$USE_CLOUDSQL_PROXY" == "auto" ]]; then
      start_proxy
      SRC_HOST="127.0.0.1"
      SRC_PORT="${SRC_PROXY_PORT}"
      SRC_SSL="disable"
    else
      err "No se pudo conectar al origen y USE_CLOUDSQL_PROXY=0."
      exit 1
    fi
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
  log "Generando dump data-only desde STAGING -> ${DUMP_FILE}"
  SRC_PASS_ENC="$(urlencode_pass "$SRC_PASS")"
  SRC_CONN="postgresql://${SRC_USER}:${SRC_PASS_ENC}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}"
  DUMP_ARGS=(-F c --no-owner --no-acl --data-only -v -f "$DUMP_FILE")
  for attempt in $(seq 1 "$PGDUMP_RETRIES"); do
    if run_pg_cmd pg_dump "$SRC_CONN" "${DUMP_ARGS[@]}"; then
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

### ===== Restore data-only (tal cual, sin cambios) =====
log "Conexión a la DB destino existente:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\conninfo"

if [[ "$TRUNCATE_BEFORE_RESTORE" == "1" ]]; then
  log "TRUNCATE de tablas public (sin schema_migrations)…"
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
DECLARE
  r record;
BEGIN
  FOR r IN
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = 'public'
      AND tablename <> 'schema_migrations'
  LOOP
    EXECUTE format('TRUNCATE TABLE public.%I RESTART IDENTITY CASCADE', r.tablename);
  END LOOP;
END $$;
SQL
fi

RESTORE_COMMON=(-h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --no-owner --no-privileges -v)
LIST_FILE="$(mktemp)"
LIST_FILE_FILTERED="$(mktemp)"

log "Generando lista filtrada (sin schema_migrations)…"
run_pg_cmd pg_restore -l "$DUMP_FILE" > "$LIST_FILE" || true
python - "$LIST_FILE" "$LIST_FILE_FILTERED" <<'PY'
import sys

src, dst = sys.argv[1:]

def should_keep(line: str) -> bool:
    if "schema_migrations" in line:
        return False
    return True

with open(src, "r", encoding="utf-8", errors="replace") as fh_in, \
     open(dst, "w", encoding="utf-8") as fh_out:
    for line in fh_in:
        if should_keep(line):
            fh_out.write(line)
PY

RESTORE_FILTERS=(--data-only --schema=public --use-list "$LIST_FILE_FILTERED")

log "RESTORE DATA-ONLY (tal cual, sin cambios)…"
if [[ "$DISABLE_TRIGGERS" == "1" ]]; then
  log "  -> con --disable-triggers"
  # Fail-fast: nunca continuar con restore parcial luego de TRUNCATE.
  PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" "${RESTORE_FILTERS[@]}" --disable-triggers "$DUMP_FILE"
else
  # Fail-fast: nunca continuar con restore parcial luego de TRUNCATE.
  PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" "${RESTORE_FILTERS[@]}" "$DUMP_FILE"
fi
log "  ✅ Data-only completado exitosamente"

rm -f "$LIST_FILE" "$LIST_FILE_FILTERED"

### ===== Verificación rápida =====
log "Tablas (primeras 30 filas de \dt):"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -c "\dt public.*" | sed -n '1,30p' || true

### ===== Fix secuencias después de data-only restore =====
log "Sincronizando secuencias con MAX(id) de cada tabla..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
DECLARE
  seq_record RECORD;
  max_id BIGINT;
BEGIN
  FOR seq_record IN
    SELECT
      seq.relname AS seq_name,
      t.relname AS table_name,
      a.attname AS column_name
    FROM pg_class seq
    JOIN pg_depend d ON d.objid = seq.oid
    JOIN pg_class t ON d.refobjid = t.oid
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
    WHERE seq.relkind = 'S'
  LOOP
    EXECUTE format('SELECT COALESCE(MAX(%I), 0) FROM %I', seq_record.column_name, seq_record.table_name) INTO max_id;
    EXECUTE format('SELECT setval(%L, %s + 1, false)', seq_record.seq_name, max_id);
  END LOOP;
END $$;
SQL
log "OK. Secuencias sincronizadas."

log "✅ RESTAURACIÓN COMPLETA (staging → local, data-only, tal cual)."
