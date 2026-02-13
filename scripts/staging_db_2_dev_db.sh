#!/usr/bin/env bash
# staging_db_2_dev_db.sh
# - Copia datos desde GCP STAGING a GCP DEV (misma instancia, distintas DBs)
# - Origen: new_ponti_db_staging. Destino: new_ponti_db_dev.
# - Tratamiento: data-only, sin schema, sin renames.
#
# Requiere: SRC_PASS (staging) y DST_PASS (dev).
# Opcional: scripts/staging_db_2_dev_db.env
#
# Uso: ./scripts/staging_db_2_dev_db.sh
#   o: cp scripts/staging_db_2_dev_db.env.example scripts/staging_db_2_dev_db.env && ./scripts/staging_db_2_dev_db.sh
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

### ===== Cargar credenciales =====
if [[ -f "${SCRIPT_DIR}/staging_db_2_dev_db.env" ]]; then
  set -a
  source "${SCRIPT_DIR}/staging_db_2_dev_db.env"
  set +a
fi

# Origen: STAGING
SRC_USER="${SRC_USER:-app_stg}"
SRC_PASS="${SRC_PASS:-}"
SRC_HOST="${SRC_HOST:-136.112.24.122}"
SRC_PORT="${SRC_PORT:-5432}"
SRC_DB="${SRC_DB:-new_ponti_db_staging}"
SRC_SSL="${SRC_SSL:-disable}"

# Destino: DEV
DST_USER="${DST_USER:-soalen-db-v3}"
DST_PASS="${DST_PASS:-}"
DST_HOST="${DST_HOST:-$SRC_HOST}"
DST_PORT="${DST_PORT:-$SRC_PORT}"
DST_DB="${DST_DB:-new_ponti_db_dev}"
DST_SSL="${DST_SSL:-$SRC_SSL}"

# Cloud SQL Proxy (fallback)
USE_CLOUDSQL_PROXY="${USE_CLOUDSQL_PROXY:-auto}"
SRC_INSTANCE_PROJECT="${SRC_INSTANCE_PROJECT:-new-ponti-dev}"
SRC_INSTANCE_REGION="${SRC_INSTANCE_REGION:-us-central1}"
SRC_INSTANCE_NAME="${SRC_INSTANCE_NAME:-new-ponti-db-dev}"
SRC_PROXY_PORT="${SRC_PROXY_PORT:-55432}"
PROXY_CONTAINER_NAME="${PROXY_CONTAINER_NAME:-ponti-cloudsql-proxy}"

DUMP_FILE="${DUMP_FILE:-/tmp/staging_to_dev_$(date +%F_%H%M%S).dump}"
TRUNCATE_BEFORE_RESTORE="${TRUNCATE_BEFORE_RESTORE:-1}"

log(){ echo -e "\n[INFO] $*"; }
warn(){ echo -e "\n[WARN] $*"; }
err(){ echo -e "\n[ERROR] $*" >&2; }
need(){ command -v "$1" >/dev/null 2>&1 || { err "No se encontró '$1' en PATH"; exit 1; }; }

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

### ===== Validaciones =====
if [[ -z "${SRC_PASS}" ]]; then
  err "SRC_PASS es requerido (staging). Creá scripts/staging_db_2_dev_db.env"
  exit 1
fi
if [[ -z "${DST_PASS}" ]]; then
  err "DST_PASS es requerido (dev). Creá scripts/staging_db_2_dev_db.env"
  exit 1
fi

need psql; need pg_dump; need pg_restore; need pg_isready

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

### ===== Conexión a origen =====
log "Chequeando acceso a origen ${SRC_HOST}:${SRC_PORT}..."
if ! PGPASSWORD="$SRC_PASS" pg_isready -h "$SRC_HOST" -p "$SRC_PORT" -U "$SRC_USER" >/dev/null 2>&1; then
  if [[ "$USE_CLOUDSQL_PROXY" == "1" || "$USE_CLOUDSQL_PROXY" == "auto" ]]; then
    start_proxy
    trap cleanup_proxy EXIT
    SRC_HOST="127.0.0.1"
    SRC_PORT="${SRC_PROXY_PORT}"
    DST_HOST="127.0.0.1"
    DST_PORT="${SRC_PROXY_PORT}"
    SRC_SSL="disable"
    DST_SSL="disable"
  else
    err "No se pudo conectar al origen. Usá USE_CLOUDSQL_PROXY=1"
    exit 1
  fi
fi

log "Origen: ${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB} (STAGING)"
log "Destino: ${DST_USER}@${DST_HOST}:${DST_PORT}/${DST_DB} (DEV)"

### ===== Dump desde STAGING =====
log "Generando dump data-only desde STAGING -> ${DUMP_FILE}"
if ! PGPASSWORD="$SRC_PASS" run_pg_cmd pg_dump \
  "postgresql://${SRC_USER}@${SRC_HOST}:${SRC_PORT}/${SRC_DB}?sslmode=${SRC_SSL}" \
  -F c --no-owner --no-acl --data-only -v -f "$DUMP_FILE"; then
  err "pg_dump falló."
  exit 1
fi

if [[ ! -s "$DUMP_FILE" ]]; then
  err "El dump está vacío: ${DUMP_FILE}"
  exit 1
fi

### ===== TRUNCATE en destino (DEV) =====
if [[ "$TRUNCATE_BEFORE_RESTORE" == "1" ]]; then
  log "TRUNCATE de tablas public en DEV (sin schema_migrations)…"
  PGPASSWORD="$DST_PASS" psql -h "$DST_HOST" -p "$DST_PORT" -U "$DST_USER" -d "$DST_DB" -v ON_ERROR_STOP=1 <<'SQL'
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

### ===== Restore data-only en DEV =====
RESTORE_COMMON=(-h "$DST_HOST" -p "$DST_PORT" -U "$DST_USER" -d "$DST_DB" --no-owner --no-privileges -v)
LIST_FILE="$(mktemp)"
LIST_FILE_FILTERED="$(mktemp)"
TABLES_FILE="$(mktemp)"
SEQS_FILE="$(mktemp)"

run_pg_cmd pg_restore -l "$DUMP_FILE" > "$LIST_FILE" || true
PGPASSWORD="$DST_PASS" psql -h "$DST_HOST" -p "$DST_PORT" -U "$DST_USER" -d "$DST_DB" -At -c \
  "SELECT tablename FROM pg_tables WHERE schemaname='public';" > "$TABLES_FILE"
PGPASSWORD="$DST_PASS" psql -h "$DST_HOST" -p "$DST_PORT" -U "$DST_USER" -d "$DST_DB" -At -c \
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

log "RESTORE data-only en DEV…"
if PGPASSWORD="$DST_PASS" run_pg_cmd pg_restore "${RESTORE_COMMON[@]}" --data-only --schema=public --use-list "$LIST_FILE_FILTERED" --disable-triggers "$DUMP_FILE"; then
  log "  ✅ Data-only completado"
else
  warn "  ⚠️  Restore tuvo errores pero continuando..."
fi

rm -f "$LIST_FILE" "$LIST_FILE_FILTERED" "$TABLES_FILE" "$SEQS_FILE"

### ===== Fix secuencias =====
log "Sincronizando secuencias..."
PGPASSWORD="$DST_PASS" psql -h "$DST_HOST" -p "$DST_PORT" -U "$DST_USER" -d "$DST_DB" -v ON_ERROR_STOP=1 <<'SQL'
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

log "✅ STAGING → DEV completado (data-only)."
log "Dump temporal: ${DUMP_FILE} (podés borrarlo si no lo necesitás)."
