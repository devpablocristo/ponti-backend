#!/usr/bin/env bash
# reset-local-db-from-prod.sh
# - Resetea la DB local, aplica migraciones y restaura datos desde PROD.
# - Origen: new_ponti_db_prod. Tratamiento: data-only, sin schema, sin renames.
# - PROD es fuente read-only: pg_dump + SELECT/introspeccion.
#
# Requiere: .env (destino local). PROD se infiere con defaults seguros y gcloud.
#
# Uso:
#   editar .env con variables reales
#   DRY_RUN=1 ./scripts/db/reset-local-db-from-prod.sh
#   ./scripts/db/reset-local-db-from-prod.sh
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CORE_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="${CORE_DIR}/.env"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "[ERROR] Falta ${ENV_FILE}."
  echo "[ERROR] Creá ${ENV_FILE} con las variables reales."
  exit 1
fi
set -a
source "${ENV_FILE}"
set +a

### ===== Origen (PROD, read-only desde este script) =====
SRC_USER="${SRC_USER:-${DB_USER_PROD:-soalen-db-v3}}"
SRC_PASS="${SRC_PASS:-}"
SRC_HOST="${SRC_HOST:-}"
SRC_PORT="${SRC_PORT:-5432}"
SRC_DB="${SRC_DB:-${DB_NAME_PROD:-new_ponti_db_prod}}"
SRC_SSL="${SRC_SSL:-disable}"

# Si no se provee SRC_PASS, intentar obtenerlo automaticamente desde Secret Manager.
# Esto evita tener que guardar contrasenas en archivos locales.
SRC_PASS_SECRET_PROJECT="${SRC_PASS_SECRET_PROJECT:-new-ponti-dev}"
SRC_PASS_SECRET_NAME="${SRC_PASS_SECRET_NAME:-db-password-dev}"
if [[ -z "${SRC_PASS}" ]] && command -v gcloud >/dev/null 2>&1; then
  # No mostrar el valor en stdout/stderr.
  if SRC_PASS="$(gcloud secrets versions access latest --secret="${SRC_PASS_SECRET_NAME}" --project="${SRC_PASS_SECRET_PROJECT}" 2>/dev/null)"; then
    :
  else
    SRC_PASS=""
  fi
fi

# Cloud SQL Proxy (fallback si no hay acceso directo)
USE_CLOUDSQL_PROXY="${USE_CLOUDSQL_PROXY:-auto}" # auto | 1 | 0
SRC_INSTANCE_PROJECT="${SRC_INSTANCE_PROJECT:-${CLOUDSQL_PROJECT_PROD:-}}"
SRC_INSTANCE_REGION="${SRC_INSTANCE_REGION:-}"
SRC_INSTANCE_NAME="${SRC_INSTANCE_NAME:-${DB_INSTANCE_NAME_PROD:-}}"
SRC_INSTANCE_CONN="${SRC_INSTANCE_CONN:-}"
SRC_PROXY_PORT="${SRC_PROXY_PORT:-55433}"
PROXY_CONTAINER_NAME="${PROXY_CONTAINER_NAME:-ponti-cloudsql-proxy}"

### ===== Destino (Local, desde .env) =====
DB_USER="${DB_USER:-}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_NAME="${DB_NAME:-}"
DB_PORT="${DB_PORT:-}"
DB_SSL_MODE="${DB_SSL_MODE:-disable}"

if [[ -z "${DB_USER}" || -z "${DB_PASSWORD}" || -z "${DB_NAME}" || -z "${DB_PORT}" ]]; then
  echo "[ERROR] En ${ENV_FILE} faltan variables del destino local."
  echo "[ERROR] Requeridas: DB_USER, DB_PASSWORD, DB_NAME, DB_PORT (DB_HOST opcional)."
  exit 1
fi

### ===== Control =====
DISABLE_TRIGGERS="${DISABLE_TRIGGERS:-1}"  # 1= intentar deshabilitar triggers (requiere superuser)
SKIP_DUMP="${SKIP_DUMP:-0}"                # 1= salta pg_dump y usa DUMP_FILE existente
DUMP_FILE="${DUMP_FILE:-/tmp/prod_to_local_$(date +%F_%H%M%S).dump}"
PGDUMP_RETRIES="${PGDUMP_RETRIES:-3}"
PGDUMP_RETRY_SLEEP="${PGDUMP_RETRY_SLEEP:-5}"
RESTORE_MODE="${RESTORE_MODE:-data-only}" # solo data-only soportado
TRUNCATE_BEFORE_RESTORE="${TRUNCATE_BEFORE_RESTORE:-1}"
RESET_LOCAL_DB="${RESET_LOCAL_DB:-1}"
ACTORS_BACKFILL_SYNC="${ACTORS_BACKFILL_SYNC:-1}"
MIGRATE_TARGET_VERSION="${MIGRATE_TARGET_VERSION:-}" # vacio => 'migrate up' (aplica TODAS las migraciones del set, sin fijar una version que podria no existir)
POST_RESTORE_TENANT_BACKFILL="${POST_RESTORE_TENANT_BACKFILL:-1}"
RUN_FINAL_MIGRATIONS="${RUN_FINAL_MIGRATIONS:-1}"
DRY_RUN="${DRY_RUN:-0}"

CRITICAL_TABLES=(
  customers
  projects
  campaigns
  fields
  lots
  workorders
  workorder_items
  work_order_drafts
  work_order_draft_items
  supply_movements
  stocks
  supplies
)

SRC_CONN=""
TEMP_FILES=()

log(){ echo -e "\n[INFO] $*"; }
warn(){ echo -e "\n[WARN] $*"; }
err(){ echo -e "\n[ERROR] $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "No se encontro '$1' en PATH"; exit 1; }; }

new_temp_file() {
  local tmp
  tmp="$(mktemp)"
  TEMP_FILES+=("$tmp")
  printf '%s\n' "$tmp"
}

# Filtra warnings conocidos sin ocultar errores de restore.
filter_pg_stderr() {
  python3 - <<'PY'
import sys
for line in sys.stdin:
    if "transaction_timeout" in line:
        continue
    if "No existing cluster is suitable as a default target" in line:
        continue
    sys.stderr.write(line)
PY
}

# URL-encode password para connection string (evita PGPASSFILE/PGPASSWORD que pueden fallar).
urlencode_pass() {
  python3 -c "
import urllib.parse
import sys
print(urllib.parse.quote(sys.argv[1], safe=''))
" "$1"
}

run_pg_cmd() {
  local tmp code=0
  tmp="$(mktemp)"
  # Captura el exit code REAL. El '|| code=$?' evita que set -e aborte aca y
  # corrige el bug previo: se tomaba $? DESPUES del 'if' (donde ya valia 0),
  # asi que la funcion SIEMPRE retornaba 0 y tragaba los errores de pg.
  "$@" 2> "$tmp" || code=$?
  filter_pg_stderr < "$tmp"
  rm -f "$tmp"
  return "${code}"
}

validate_destination_safety() {
  log "Validando guardas de seguridad del destino local..."

  case "${DB_HOST}" in
    localhost|127.0.0.1) ;;
    *)
      err "destino bloqueado: DB_HOST debe ser localhost/127.0.0.1; recibido '${DB_HOST}'"
      exit 1
      ;;
  esac

  if [[ ! "${DB_PORT}" =~ ^[0-9]+$ ]]; then
    err "destino bloqueado: DB_PORT debe ser numerico; recibido '${DB_PORT}'"
    exit 1
  fi

  if [[ ! "${DB_NAME}" =~ ^[A-Za-z0-9_]+$ ]]; then
    err "destino bloqueado: DB_NAME solo puede contener letras, numeros y underscore; recibido '${DB_NAME}'"
    exit 1
  fi

  local db_name_lower="${DB_NAME,,}"
  local risky_patterns=(production prod prd live)
  local pattern
  for pattern in "${risky_patterns[@]}"; do
    if [[ "${db_name_lower}" == *"${pattern}"* ]]; then
      err "destino bloqueado: DB_NAME='${DB_NAME}' contiene patron productivo: ${pattern}"
      exit 1
    fi
  done

  if [[ "${RESTORE_MODE}" != "data-only" ]]; then
    err "RESTORE_MODE='${RESTORE_MODE}' no soportado. Este script solo permite RESTORE_MODE=data-only."
    exit 1
  fi

  if [[ "${RESET_LOCAL_DB}" != "1" ]]; then
    err "RESET_LOCAL_DB=0 no soportado para un restore PRD -> local verificable. El flujo exige reset total."
    exit 1
  fi

  log "OK: destino permitido ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
}

validate_required_tools() {
  log "Validando dependencias locales..."
  need python3
  need psql
  need pg_dump
  need pg_restore
  need pg_isready
  if [[ "${RESET_LOCAL_DB}" == "1" || "${RUN_FINAL_MIGRATIONS}" == "1" || "${USE_CLOUDSQL_PROXY}" != "0" ]]; then
    need docker
  fi
  log "OK: dependencias disponibles"
}

infer_src_host_from_gcloud() {
  if ! command -v gcloud >/dev/null 2>&1; then
    return 1
  fi

  # Defaults seguros para este repo: PROD vive en la instancia unificada de DEV.
  local proj="${SRC_INSTANCE_PROJECT:-new-ponti-dev}"
  local inst="${SRC_INSTANCE_NAME:-${DB_INSTANCE_NAME_DEV:-new-ponti-db-dev}}"

  local ip
  ip="$(gcloud sql instances describe "$inst" --project="$proj" --format='value(ipAddresses[0].ipAddress)' 2>/dev/null | tr -d '\r' || true)"
  if [[ -n "${ip}" ]]; then
    SRC_HOST="${ip}"
    return 0
  fi
  return 1
}

cleanup_proxy() {
  if command -v docker >/dev/null 2>&1 && docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${PROXY_CONTAINER_NAME}$"; then
    docker rm -f "${PROXY_CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi
}

cleanup_all() {
  cleanup_proxy
  local tmp
  for tmp in "${TEMP_FILES[@]}"; do
    [[ -n "${tmp}" ]] && rm -f "${tmp}" || true
  done
}

trap cleanup_all EXIT

start_proxy() {
  # Solo reutilizar si NUESTRO contenedor proxy esta corriendo.
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${PROXY_CONTAINER_NAME}$"; then
    if PGPASSWORD="${SRC_PASS}" pg_isready -h "127.0.0.1" -p "${SRC_PROXY_PORT}" -U "${SRC_USER}" >/dev/null 2>&1; then
      log "Proxy ya activo (${PROXY_CONTAINER_NAME}), reutilizando"
      return 0
    fi
  fi

  if [[ -z "${SRC_INSTANCE_CONN}" ]]; then
    if [[ -z "${SRC_INSTANCE_PROJECT}" && -z "${SRC_INSTANCE_REGION}" && -z "${SRC_INSTANCE_NAME}" ]]; then
      err "Para usar Cloud SQL Proxy faltan SRC_INSTANCE_* (project/region/name) o SRC_INSTANCE_CONN."
      err "Definilos en .env o en el entorno antes de ejecutar."
      return 1
    fi
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

  for _ in {1..15}; do
    if PGPASSWORD="${SRC_PASS}" pg_isready -h "127.0.0.1" -p "${SRC_PROXY_PORT}" -U "${SRC_USER}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  err "Cloud SQL Proxy no respondio. Verifica credenciales ADC o permisos cloudsql.client."
  return 1
}

resolve_source_connection() {
  log "Validando credenciales y conexion de origen PROD..."

  if [[ -z "${SRC_PASS}" ]]; then
    err "SRC_PASS es requerido para el usuario de PROD (${SRC_USER})."
    err "Opciones:"
    err "  - Setea SRC_PASS en el entorno antes de ejecutar"
    err "  - O configura gcloud ADC y deja que el script lea db-password-dev desde Secret Manager"
    exit 1
  fi

  if [[ -z "${SRC_USER}" || -z "${SRC_DB}" || -z "${SRC_PORT}" ]]; then
    err "Faltan credenciales minimas de origen. Defini SRC_USER, SRC_PASS, SRC_DB, SRC_PORT."
    err "Ejemplo: SRC_PASS='...' ./scripts/db/reset-local-db-from-prod.sh"
    exit 1
  fi

  if [[ -z "${SRC_HOST}" ]]; then
    infer_src_host_from_gcloud || true
  fi

  if [[ "${USE_CLOUDSQL_PROXY}" == "1" || -z "${SRC_HOST}" ]]; then
    if [[ "${USE_CLOUDSQL_PROXY}" == "0" ]]; then
      err "SRC_HOST vacio y USE_CLOUDSQL_PROXY=0. Defini SRC_HOST o usa USE_CLOUDSQL_PROXY=auto/1."
      exit 1
    fi
    log "Usando Cloud SQL Proxy para origen PROD"
    start_proxy
    SRC_HOST="127.0.0.1"
    SRC_PORT="${SRC_PROXY_PORT}"
    SRC_SSL="disable"
  else
    log "Chequeando acceso directo a origen ${SRC_HOST}:${SRC_PORT}..."
    if ! PGPASSWORD="${SRC_PASS}" pg_isready -h "${SRC_HOST}" -p "${SRC_PORT}" -U "${SRC_USER}" >/dev/null 2>&1; then
      if [[ "${USE_CLOUDSQL_PROXY}" == "auto" ]]; then
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

  if [[ -z "${SRC_SSL}" ]]; then
    SRC_SSL="disable"
  fi

  local src_pass_enc
  src_pass_enc="$(urlencode_pass "${SRC_PASS}")"
  SRC_CONN="postgresql://${SRC_USER}:${src_pass_enc}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}"

  # SELECT read-only para validar login real, sin writes sobre PROD.
  run_src_psql_readonly -c "SELECT current_database();" >/dev/null
  log "OK: origen efectivo ${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB} (PROD read-only, sslmode=${SRC_SSL})"
}

validate_destination_connection() {
  log "Esperando PostgreSQL local en ${DB_HOST}:${DB_PORT} como ${DB_USER}..."
  for i in {1..30}; do
    if PGPASSWORD="${DB_PASSWORD}" pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" >/dev/null 2>&1; then
      break
    fi
    sleep 1
    if [[ "$i" -eq 30 ]]; then
      err "PG local no responde o credenciales invalidas.
Solucion rapida:
  - Nativo: sudo -u postgres psql -c \"ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';\"
  - Docker: docker exec -it <contenedor> psql -U postgres -c \"ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';\""
      exit 1
    fi
  done

  run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d postgres -v ON_ERROR_STOP=1 -X -c "\conninfo" >/dev/null
  log "OK: conexion local validada"
}

run_src_psql_readonly() {
  run_pg_cmd env PGOPTIONS="-c default_transaction_read_only=on" \
    psql -v ON_ERROR_STOP=1 -X -At "$@" "${SRC_CONN}"
}

run_dest_psql() {
  run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" \
    psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 -X "$@"
}

run_dest_postgres_psql() {
  run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" \
    psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d postgres -v ON_ERROR_STOP=1 -X "$@"
}

capture_critical_counts() {
  local label="$1"
  local mode="$2"
  local output_file="$3"
  local missing=0
  local table exists count

  : > "${output_file}"
  log "Contando tablas criticas en ${label}..."

  for table in "${CRITICAL_TABLES[@]}"; do
    if [[ "${mode}" == "source" ]]; then
      exists="$(run_src_psql_readonly -c "SELECT CASE WHEN to_regclass('public.${table}') IS NULL THEN 'missing' ELSE 'present' END;")"
    else
      exists="$(run_dest_psql -At -c "SELECT CASE WHEN to_regclass('public.${table}') IS NULL THEN 'missing' ELSE 'present' END;")"
    fi

    if [[ "${exists}" != "present" ]]; then
      printf '%s\tMISSING\n' "${table}" >> "${output_file}"
      err "tabla critica faltante en ${label}: public.${table}"
      missing=1
      continue
    fi

    if [[ "${mode}" == "source" ]]; then
      count="$(run_src_psql_readonly -c "SELECT COUNT(*) FROM public.${table};")"
    else
      count="$(run_dest_psql -At -c "SELECT COUNT(*) FROM public.${table};")"
    fi
    printf '%s\t%s\n' "${table}" "${count}" >> "${output_file}"
    log "  - ${table}: ${count}"
  done

  return "${missing}"
}

count_from_file() {
  local counts_file="$1"
  local table="$2"
  awk -F '\t' -v table="${table}" '$1 == table { print $2 }' "${counts_file}"
}

warn_operational_data_gaps() {
  local source_counts_file="$1"
  local workorder_items_count
  workorder_items_count="$(count_from_file "${source_counts_file}" "workorder_items")"

  if [[ "${workorder_items_count}" == "0" ]]; then
    warn "======================================================================"
    warn "WARN CRITICO: PROD tiene workorder_items=0."
    warn "El restore local sera fiel, pero dashboards con insumos ejecutados"
    warn "no podran reproducirse desde PROD porque esos datos no existen en origen."
    warn "Esto NO es fallo de restore. Es un gap de datos en PROD."
    warn "======================================================================"
  fi
}

validate_dump_manifest() {
  local dump_file="$1"
  local manifest_file="$2"
  local missing=0
  local table

  log "Validando manifest del dump para tablas criticas..."
  run_pg_cmd pg_restore -l "${dump_file}" > "${manifest_file}"

  for table in "${CRITICAL_TABLES[@]}"; do
    if ! grep -Eq " TABLE DATA public ${table} " "${manifest_file}"; then
      err "dump incompleto: falta TABLE DATA public.${table}"
      missing=1
    fi
  done

  if [[ "${missing}" == "1" ]]; then
    err "Abortando antes de tocar local: el dump no contiene todas las tablas criticas."
    exit 1
  fi

  log "OK: dump contiene TABLE DATA para todas las tablas criticas"
}

compare_critical_counts() {
  local source_counts_file="$1"
  local dest_counts_file="$2"
  local failed=0
  local table source_count dest_count

  log "Comparando conteos criticos PROD vs local..."
  for table in "${CRITICAL_TABLES[@]}"; do
    source_count="$(count_from_file "${source_counts_file}" "${table}")"
    dest_count="$(count_from_file "${dest_counts_file}" "${table}")"

    if [[ -z "${source_count}" || "${source_count}" == "MISSING" ]]; then
      err "conteo critico faltante en PROD: ${table} source=${source_count:-MISSING}"
      failed=1
      continue
    fi

    if [[ -z "${dest_count}" || "${dest_count}" == "MISSING" ]]; then
      err "conteo critico faltante en local: ${table} local=${dest_count:-MISSING}"
      failed=1
      continue
    fi

    if [[ "${source_count}" != "${dest_count}" ]]; then
      err "conteo critico no coincide: ${table} source=${source_count} local=${dest_count}"
      failed=1
    else
      log "  - ${table}: OK (${source_count})"
    fi
  done

  if [[ "${failed}" == "1" ]]; then
    err "Restore abortado: los conteos criticos no coinciden."
    exit 1
  fi

  log "OK: conteos criticos PROD/local coinciden. PROD fue usado solo como fuente read-only."
}

print_dry_run_plan() {
  log "DRY_RUN=1: no se ejecutara dump, reset, restore, migraciones ni backfills."
  cat <<EOF

[DRY-RUN] Flujo que ejecutaria:
  1. pg_dump data-only desde PROD read-only -> ${DUMP_FILE}
  2. validar manifest del dump
  3. reset TOTAL de DB local ${DB_NAME}
  4. crear DB limpia y migrar hasta version ${MIGRATE_TARGET_VERSION}
  5. restore data-only de schema public
  6. post-restore tenant backfill=${POST_RESTORE_TENANT_BACKFILL}
  7. migraciones finales=${RUN_FINAL_MIGRATIONS}
  8. actors backfill sync=${ACTORS_BACKFILL_SYNC}
  9. sincronizar secuencias
  10. comparar conteos criticos PROD/local

[DRY-RUN] Garantias:
  - Origen PROD: solo pg_dump + SELECT/introspeccion con default_transaction_read_only=on.
  - Destino local: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}
  - Patrones productivos bloqueados en DB_NAME: production, prod, prd, live.
EOF
}

generate_or_reuse_dump() {
  if [[ "${SKIP_DUMP}" == "1" ]]; then
    if [[ ! -f "${DUMP_FILE}" ]]; then
      err "SKIP_DUMP=1 pero DUMP_FILE no existe: ${DUMP_FILE}"
      exit 1
    fi
    log "SKIP_DUMP=1 -> uso dump existente: ${DUMP_FILE}"
  else
    log "Generando dump data-only desde PROD read-only -> ${DUMP_FILE}"
    local dump_args=(-F c --no-owner --no-acl --data-only -v -f "${DUMP_FILE}")
    local attempt
    for attempt in $(seq 1 "${PGDUMP_RETRIES}"); do
      if run_pg_cmd env PGOPTIONS="-c default_transaction_read_only=on" pg_dump "${SRC_CONN}" "${dump_args[@]}"; then
        break
      fi
      if [[ "${attempt}" -lt "${PGDUMP_RETRIES}" ]]; then
        warn "pg_dump fallo (intento ${attempt}/${PGDUMP_RETRIES}). Reintento en ${PGDUMP_RETRY_SLEEP}s..."
        sleep "${PGDUMP_RETRY_SLEEP}"
      else
        err "pg_dump fallo luego de ${PGDUMP_RETRIES} intentos."
        exit 1
      fi
    done
  fi

  if [[ ! -s "${DUMP_FILE}" ]]; then
    err "El dump esta vacio o incompleto: ${DUMP_FILE}"
    exit 1
  fi
}

reset_local_database() {
  log "Reset TOTAL de DB local ${DB_NAME}..."
  run_dest_postgres_psql -v db_name="${DB_NAME}" <<'SQL'
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = :'db_name'
  AND pid <> pg_backend_pid();

DROP DATABASE IF EXISTS :"db_name";
CREATE DATABASE :"db_name";
SQL
}

run_local_migrations_to_target() {
  log "Aplicando migraciones locales..."
  local db_password_enc
  db_password_enc="$(urlencode_pass "${DB_PASSWORD}")"
  local migrate_common=(
    docker run --rm --network host
    -v "${CORE_DIR}/migrations_v4:/migrations:ro"
    migrate/migrate:v4.17.1
    -path /migrations
    -database "postgres://${DB_USER}:${db_password_enc}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}"
  )

  if [[ -n "${MIGRATE_TARGET_VERSION}" ]]; then
    log "  -> hasta version ${MIGRATE_TARGET_VERSION}"
    "${migrate_common[@]}" goto "${MIGRATE_TARGET_VERSION}"
  else
    "${migrate_common[@]}" up
  fi
}

detect_trigger_strategy() {
  local is_super
  is_super="$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -At -d postgres -c "SELECT rolsuper::int FROM pg_roles WHERE rolname='${DB_USER}';" || echo 0)"
  if [[ "${is_super:-0}" != "1" && "${DISABLE_TRIGGERS}" == "1" ]]; then
    warn "El rol '${DB_USER}' NO es superuser. Cambio automatico DISABLE_TRIGGERS=0 para evitar errores."
    DISABLE_TRIGGERS=0
  fi
}

truncate_public_tables() {
  if [[ "${TRUNCATE_BEFORE_RESTORE}" != "1" ]]; then
    return 0
  fi

  log "TRUNCATE de tablas public (sin schema_migrations)..."
  run_dest_psql <<'SQL'
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
}

restore_data_only() {
  local list_file list_file_filtered existing_relations_file skipped_entries_file
  list_file="$(new_temp_file)"
  list_file_filtered="$(new_temp_file)"
  existing_relations_file="$(new_temp_file)"
  skipped_entries_file="$(new_temp_file)"

  log "Generando lista filtrada del dump (sin schema_migrations y sin tablas fuera del schema local)..."
  run_pg_cmd pg_restore -l "${DUMP_FILE}" > "${list_file}"
  run_dest_psql -At -c "SELECT c.relname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'S') ORDER BY c.relname;" > "${existing_relations_file}"

  python3 - "${list_file}" "${list_file_filtered}" "${existing_relations_file}" "${skipped_entries_file}" <<'PY'
import sys

src, dst, existing_path, skipped_path = sys.argv[1:]

with open(existing_path, "r", encoding="utf-8", errors="replace") as fh:
    existing = {line.strip() for line in fh if line.strip()}

def entry_name(line: str, marker: str):
    if marker not in line:
        return None
    rest = line.split(marker, 1)[1].strip()
    if not rest:
        return None
    return rest.split()[0]

def should_keep(line: str) -> bool:
    if "schema_migrations" in line:
        return False
    table = entry_name(line, " TABLE DATA public ")
    if table is not None and table not in existing:
        return False
    sequence = entry_name(line, " SEQUENCE SET public ")
    if sequence is not None and sequence not in existing:
        return False
    return True

with open(src, "r", encoding="utf-8", errors="replace") as fh_in, \
     open(dst, "w", encoding="utf-8") as fh_out, \
     open(skipped_path, "w", encoding="utf-8") as fh_skipped:
    for line in fh_in:
        if should_keep(line):
            fh_out.write(line)
        elif " TABLE DATA public " in line or " SEQUENCE SET public " in line:
            fh_skipped.write(line)
PY

  if [[ -s "${skipped_entries_file}" ]]; then
    warn "Entradas omitidas antes del restore (schema_migrations o relaciones ausentes en el schema local migrado):"
    sed -n '1,40s/^/[WARN]   - /p' "${skipped_entries_file}"
  fi

  local table
  for table in "${CRITICAL_TABLES[@]}"; do
    if ! grep -Eq " TABLE DATA public ${table} " "${list_file_filtered}"; then
      err "restore filtrado invalido: falta TABLE DATA public.${table} en la lista final"
      exit 1
    fi
  done

  local restore_common=(-h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" --no-owner --no-privileges --exit-on-error -v)
  local restore_filters=(--data-only --schema=public --use-list "${list_file_filtered}")

  log "RESTORE DATA-ONLY (schema public, sin schema_migrations)..."
  if [[ "${DISABLE_TRIGGERS}" == "1" ]]; then
    log "  -> con --disable-triggers"
    run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" pg_restore "${restore_common[@]}" "${restore_filters[@]}" --disable-triggers "${DUMP_FILE}"
  else
    run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" pg_restore "${restore_common[@]}" "${restore_filters[@]}" "${DUMP_FILE}"
  fi
  log "OK: data-only completado exitosamente"
}

run_post_restore_steps() {
  local tenant_sql="${CORE_DIR}/migrations_v4/000224_tenant_security_foundation.up.sql"
  if [[ "${POST_RESTORE_TENANT_BACKFILL}" == "1" && -f "${tenant_sql}" ]]; then
    log "Reejecutando backfill tenant post-restore sobre local..."
    run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 -f "${tenant_sql}" \
      || warn "backfill tenant fallo (paso best-effort post-restore); continuo."
  elif [[ "${POST_RESTORE_TENANT_BACKFILL}" == "1" ]]; then
    log "Skip backfill tenant: no existe ${tenant_sql} (este set de migraciones no lo incluye)."
  fi

  if [[ "${RUN_FINAL_MIGRATIONS}" == "1" ]]; then
    log "Aplicando migraciones finales post-restore sobre local..."
    local db_password_enc
    db_password_enc="$(urlencode_pass "${DB_PASSWORD}")"
    docker run --rm --network host \
      -v "${CORE_DIR}/migrations_v4:/migrations:ro" \
      migrate/migrate:v4.17.1 \
      -path /migrations \
      -database "postgres://${DB_USER}:${db_password_enc}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
      up
  fi

  if [[ "${ACTORS_BACKFILL_SYNC}" == "1" ]]; then
    log "Reejecutando backfill/sync de actors sobre local..."
    run_pg_cmd env PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 -f "${CORE_DIR}/scripts/db/actors_backfill_sync.sql" \
      || warn "backfill actors fallo (esquema sin tabla actors?, paso best-effort); continuo."
  fi
}

sync_sequences() {
  log "Sincronizando secuencias con MAX(id) de cada tabla..."
  run_dest_psql <<'SQL'
DO $$
DECLARE
  seq_record RECORD;
  max_id BIGINT;
BEGIN
  FOR seq_record IN
    SELECT
      seq.relname AS seq_name,
      n.nspname AS table_schema,
      t.relname AS table_name,
      a.attname AS column_name
    FROM pg_class seq
    JOIN pg_depend d ON d.objid = seq.oid
    JOIN pg_class t ON d.refobjid = t.oid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
    WHERE seq.relkind = 'S'
  LOOP
    EXECUTE format('SELECT COALESCE(MAX(%I), 0) FROM %I.%I', seq_record.column_name, seq_record.table_schema, seq_record.table_name) INTO max_id;
    EXECUTE format('SELECT setval(%L, %s + 1, false)', seq_record.seq_name, max_id);
  END LOOP;
END $$;
SQL
  log "OK: secuencias sincronizadas."
}

### ===== Main =====
validate_destination_safety
validate_required_tools
resolve_source_connection
validate_destination_connection

SOURCE_COUNTS_FILE="$(new_temp_file)"
SOURCE_COUNTS_MISSING=0
capture_critical_counts "PROD" "source" "${SOURCE_COUNTS_FILE}" || SOURCE_COUNTS_MISSING=1
if [[ "${SOURCE_COUNTS_MISSING}" == "1" ]]; then
  err "Abortando: PROD no contiene todas las tablas criticas esperadas."
  exit 1
fi
warn_operational_data_gaps "${SOURCE_COUNTS_FILE}"

if [[ "${DRY_RUN}" == "1" ]]; then
  print_dry_run_plan
  if [[ "${SKIP_DUMP}" == "1" && -f "${DUMP_FILE}" ]]; then
    MANIFEST_FILE="$(new_temp_file)"
    validate_dump_manifest "${DUMP_FILE}" "${MANIFEST_FILE}"
  fi
  log "DRY_RUN finalizado OK. No se destruyo ni restauro nada."
  exit 0
fi

generate_or_reuse_dump

MANIFEST_FILE="$(new_temp_file)"
validate_dump_manifest "${DUMP_FILE}" "${MANIFEST_FILE}"

log "Contenido del dump (primeras 20 lineas):"
run_pg_cmd pg_restore -l "${DUMP_FILE}" | head -n 20

reset_local_database
run_local_migrations_to_target
detect_trigger_strategy
truncate_public_tables
restore_data_only
run_post_restore_steps
sync_sequences

log "Tablas locales (primeras 30 filas de \\dt):"
run_dest_psql -c "\dt public.*" | sed -n '1,30p' || true

DEST_COUNTS_FILE="$(new_temp_file)"
if ! capture_critical_counts "local" "dest" "${DEST_COUNTS_FILE}"; then
  warn "Hay tablas criticas faltantes en local; la comparacion final va a fallar con detalle."
fi
compare_critical_counts "${SOURCE_COUNTS_FILE}" "${DEST_COUNTS_FILE}"

log "RESTAURACION COMPLETA (PROD -> local, data-only, verificada)."
