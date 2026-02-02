#!/usr/bin/env bash
# new-download-gcp-db.sh
# - Descarga dump desde GCP y restaura datos en DB local v4-only
# - Aplica rename local app_parameters -> business_parameters si aplica
set -Eeuo pipefail

### ===== Preservar destino local antes de cargar .env =====
LOCAL_DB_USER="${DB_USER:-}"
LOCAL_DB_PASSWORD="${DB_PASSWORD:-}"
LOCAL_DB_HOST="${DB_HOST:-}"
LOCAL_DB_NAME="${DB_NAME:-}"
LOCAL_DB_PORT="${DB_PORT:-}"

### ===== No cargar archivos .env por ambiente =====

### ===== Origen (GCP DEV) =====
# Defaults apuntan a la DB del servicio Cloud Run (override por variables de entorno)
SRC_USER="${SRC_USER:-${DB_USER:-}}"
SRC_PASS="${SRC_PASS:-${DB_PASSWORD:-}}"
SRC_HOST="${SRC_HOST:-${DB_HOST:-}}"
SRC_DB="${SRC_DB:-${DB_NAME:-}}"
SRC_PORT="${SRC_PORT:-${DB_PORT:-}}"
SRC_SSL="${SRC_SSL:-${DB_SSL_MODE:-}}"    # disable | require | verify-full

# Opcional: tomar datos desde un servicio Cloud Run (usa gcloud si está disponible)
SRC_FROM_CLOUD_RUN="${SRC_FROM_CLOUD_RUN:-1}" # 1=leer servicio; 0=no
SRC_FORCE_CLOUD_RUN="${SRC_FORCE_CLOUD_RUN:-1}" # 1=sobrescribe con Cloud Run
SRC_SERVICE_NAME="${SRC_SERVICE_NAME:-ponti-backend}"
SRC_PROJECT_ID="${SRC_PROJECT_ID:-new-ponti-dev}"
SRC_REGION="${SRC_REGION:-us-central1}"

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
RESTORE_MODE="${RESTORE_MODE:-data-only}" # data-only | full
TRUNCATE_BEFORE_RESTORE="${TRUNCATE_BEFORE_RESTORE:-1}"
RECONCILE_CUSTOMER_ARCHIVE="${RECONCILE_CUSTOMER_ARCHIVE:-1}"

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

### ===== Completar origen desde Cloud Run si falta info =====
if [[ "${SRC_FROM_CLOUD_RUN}" == "1" && -x "$(command -v gcloud)" ]]; then
  log "Leyendo configuración desde Cloud Run: ${SRC_SERVICE_NAME} (${SRC_PROJECT_ID}/${SRC_REGION})..."
  service_json="$(gcloud run services describe "${SRC_SERVICE_NAME}" --project="${SRC_PROJECT_ID}" --region="${SRC_REGION}" --format=json 2>/dev/null || true)"
  if [[ -n "${service_json}" ]]; then
    if [[ "${SRC_FORCE_CLOUD_RUN}" == "1" ]]; then
      SRC_HOST="$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_HOST"), ""))' <<<"${service_json}")"
      SRC_USER="$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_USER"), ""))' <<<"${service_json}")"
      SRC_PASS="$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_PASSWORD"), ""))' <<<"${service_json}")"
      SRC_DB="$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_NAME"), ""))' <<<"${service_json}")"
      SRC_PORT="$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_PORT"), ""))' <<<"${service_json}")"
      SRC_SSL="$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_SSL_MODE"), ""))' <<<"${service_json}")"
    else
      SRC_HOST="${SRC_HOST:-$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_HOST"), ""))' <<<"${service_json}")}"
      SRC_USER="${SRC_USER:-$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_USER"), ""))' <<<"${service_json}")}"
      SRC_PASS="${SRC_PASS:-$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_PASSWORD"), ""))' <<<"${service_json}")}"
      SRC_DB="${SRC_DB:-$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_NAME"), ""))' <<<"${service_json}")}"
      SRC_PORT="${SRC_PORT:-$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_PORT"), ""))' <<<"${service_json}")}"
      SRC_SSL="${SRC_SSL:-$(python -c 'import json,sys; data=json.load(sys.stdin); env=data["spec"]["template"]["spec"]["containers"][0].get("env", []); print(next((item.get("value","") for item in env if item.get("name")=="DB_SSL_MODE"), ""))' <<<"${service_json}")}"
    fi
  fi
fi

### ===== Validaciones de credenciales origen =====
if [[ -z "${SRC_USER}" || -z "${SRC_PASS}" || -z "${SRC_HOST}" || -z "${SRC_DB}" || -z "${SRC_PORT}" ]]; then
  err "Faltan credenciales de origen. Definí SRC_USER, SRC_PASS, SRC_HOST, SRC_DB, SRC_PORT."
  err "Ejemplo: SRC_USER=... SRC_PASS=... SRC_HOST=... SRC_DB=... SRC_PORT=... ./new-download-gcp-db.sh"
  exit 1
fi

if [[ -z "${SRC_SSL}" ]]; then
  SRC_SSL="disable"
fi

log "Origen efectivo: ${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB} (sslmode=${SRC_SSL})"

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
  # Si el restore es data-only, generamos un dump de datos únicamente
  DUMP_ARGS=(-F c --no-owner --no-acl -v -f "$DUMP_FILE")
  if [[ "${RESTORE_MODE}" == "data-only" ]]; then
    DUMP_ARGS+=(--data-only)
  fi
  for attempt in $(seq 1 "$PGDUMP_RETRIES"); do
    if PGPASSWORD="$SRC_PASS" run_pg_cmd pg_dump \
      "postgresql://${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}" \
      "${DUMP_ARGS[@]}"; then
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

if [[ "${RESTORE_MODE}" == "full" ]]; then
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
else
  ### ===== Restore data-only (v4-only en schema existente) =====
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

  log "Preparando tabla temporal app_parameters si aplica…"
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
BEGIN
  IF to_regclass('public.business_parameters') IS NOT NULL
     AND to_regclass('public.app_parameters') IS NULL THEN
    CREATE TABLE public.app_parameters (LIKE public.business_parameters INCLUDING ALL);
  END IF;
END $$;
SQL

  RESTORE_COMMON=(-h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --no-owner --no-privileges -v)
  LIST_FILE="$(mktemp)"
  LIST_FILE_FILTERED="$(mktemp)"
  TABLES_FILE="$(mktemp)"
  SEQS_FILE="$(mktemp)"

  log "Generando lista filtrada (sin schema_migrations)…"
  run_pg_cmd pg_restore -l "$DUMP_FILE" > "$LIST_FILE" || true
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -At -c \
    "SELECT tablename FROM pg_tables WHERE schemaname='public';" > "$TABLES_FILE"
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -At -c \
    "SELECT sequencename FROM pg_sequences WHERE schemaname='public';" > "$SEQS_FILE"
  python - "$LIST_FILE" "$LIST_FILE_FILTERED" "$TABLES_FILE" "$SEQS_FILE" <<'PY'
import sys

src, dst, tables_path, seqs_path = sys.argv[1:]
with open(tables_path, "r", encoding="utf-8", errors="replace") as fh:
    tables = {line.strip() for line in fh if line.strip()}
with open(seqs_path, "r", encoding="utf-8", errors="replace") as fh:
    seqs = {line.strip() for line in fh if line.strip()}

def should_keep(line: str) -> bool:
    if "schema_migrations" in line:
        return False
    stripped = line.strip()
    if not stripped or stripped.startswith(";"):
        return True
    try:
        _, rest = line.split(";", 1)
    except ValueError:
        return True
    tokens = rest.strip().split()
    if len(tokens) < 5:
        return True
    if len(tokens) >= 6 and tokens[2] == "TABLE" and tokens[3] == "DATA":
        return tokens[5] in tables
    if len(tokens) >= 5 and tokens[2] == "SEQUENCE" and tokens[3] == "SET":
        return tokens[4] in seqs
    return True

with open(src, "r", encoding="utf-8", errors="replace") as fh_in, \
     open(dst, "w", encoding="utf-8") as fh_out:
    for line in fh_in:
        if should_keep(line):
            fh_out.write(line)
PY

  RESTORE_FILTERS=(--data-only --schema=public --use-list "$LIST_FILE_FILTERED")

  log "RESTORE DATA-ONLY (schema public)…"
  if [[ "$DISABLE_TRIGGERS" == "1" ]]; then
    log "  -> con --disable-triggers (sin --single-transaction para evitar transaction_timeout)"
    if PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" "${RESTORE_FILTERS[@]}" --disable-triggers "$DUMP_FILE"; then
      log "  ✅ Data-only completado exitosamente"
    else
      warn "  ⚠️  Data-only tuvo errores pero continuando..."
    fi
  else
    log "  -> SIN --disable-triggers (sin --single-transaction para evitar transaction_timeout)"
    if PGPASSWORD="$DB_PASSWORD" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" "${RESTORE_FILTERS[@]}" "$DUMP_FILE"; then
      log "  ✅ Data-only completado exitosamente"
    else
      warn "  ⚠️  Data-only tuvo errores pero continuando..."
    fi
  fi

  rm -f "$LIST_FILE" "$LIST_FILE_FILTERED" "$TABLES_FILE" "$SEQS_FILE"

  log "Normalizando app_parameters -> business_parameters (solo datos)…"
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
BEGIN
  IF to_regclass('public.app_parameters') IS NOT NULL THEN
    INSERT INTO public.business_parameters (id, key, value, type, category, description, created_at, updated_at)
    SELECT id, key, value, type, category, description, created_at, updated_at
    FROM public.app_parameters;
    DROP TABLE public.app_parameters;
  END IF;
END $$;
SQL
fi

### ===== Verificación rápida =====
log "Tablas (primeras 30 filas de \dt):"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 -c "\dt public.*" | sed -n '1,30p' || true

### ===== Rename local final =====
log "Aplicando rename local a business_parameters (solo DB local)..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
BEGIN
  IF to_regclass('public.app_parameters') IS NOT NULL THEN
    ALTER TABLE public.app_parameters RENAME TO business_parameters;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.app_parameters_id_seq') IS NOT NULL THEN
    ALTER SEQUENCE public.app_parameters_id_seq RENAME TO business_parameters_id_seq;
  END IF;
END $$;

ALTER SEQUENCE IF EXISTS public.business_parameters_id_seq OWNED BY public.business_parameters.id;
ALTER TABLE IF EXISTS public.business_parameters
  ALTER COLUMN id SET DEFAULT nextval('public.business_parameters_id_seq'::regclass);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'app_parameters_pkey'
      AND conrelid = 'public.business_parameters'::regclass
  ) THEN
    ALTER TABLE public.business_parameters
      RENAME CONSTRAINT app_parameters_pkey TO business_parameters_pkey;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'app_parameters_key_key'
      AND conrelid = 'public.business_parameters'::regclass
  ) THEN
    ALTER TABLE public.business_parameters
      RENAME CONSTRAINT app_parameters_key_key TO business_parameters_key_key;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.idx_app_parameters_key') IS NOT NULL THEN
    ALTER INDEX public.idx_app_parameters_key RENAME TO idx_business_parameters_key;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.idx_app_parameters_category') IS NOT NULL THEN
    ALTER INDEX public.idx_app_parameters_category RENAME TO idx_business_parameters_category;
  END IF;
END $$;

DROP FUNCTION IF EXISTS public.get_app_parameter(varchar);
DROP FUNCTION IF EXISTS public.get_app_parameter_decimal(varchar);
DROP FUNCTION IF EXISTS public.get_app_parameter_integer(varchar);

CREATE OR REPLACE FUNCTION public.get_business_parameter(p_key varchar)
RETURNS varchar AS $$
BEGIN
  RETURN (SELECT value FROM public.business_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_business_parameter_decimal(p_key varchar)
RETURNS decimal AS $$
BEGIN
  RETURN (SELECT value::decimal FROM public.business_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_business_parameter_integer(p_key varchar)
RETURNS integer AS $$
BEGIN
  RETURN (SELECT value::integer FROM public.business_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_iva_percentage()
RETURNS decimal AS $$
BEGIN
  RETURN public.get_business_parameter_decimal('iva_percentage');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_campaign_closure_days()
RETURNS integer AS $$
BEGIN
  RETURN public.get_business_parameter_integer('campaign_closure_days');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_default_fx_rate()
RETURNS decimal AS $$
BEGIN
  RETURN public.get_business_parameter_decimal('default_fx_rate');
END;
$$ LANGUAGE plpgsql IMMUTABLE;
SQL

log "OK. Renombrado aplicado en DB local."

if [[ "${RECONCILE_CUSTOMER_ARCHIVE}" == "1" ]]; then
  log "Reconciliando estado de clientes (soft delete) según proyectos activos..."
  PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -v ON_ERROR_STOP=1 -f "${ROOT_DIR}/scripts/db/db_reconcile_customer_archive.sql"
  log "OK. Reconciliación de clientes aplicada."
fi
