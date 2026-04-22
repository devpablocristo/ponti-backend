#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "Uso: $0 <source_db> <target_db>" >&2
  exit 1
fi

SOURCE_DB="$1"
TARGET_DB="$2"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DUMP_FILE="/tmp/${SOURCE_DB}_to_${TARGET_DB}_$(date +%s).dump"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

echo "Copiando ${SOURCE_DB} -> ${TARGET_DB}..."
"${ROOT_DIR}/scripts/compose_with_env.sh" up -d ponti-db >/dev/null

for i in {1..30}; do
  if "${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db pg_isready -U "${DB_USER}" -d postgres -p 5432 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db psql -U "${DB_USER}" -d postgres -v ON_ERROR_STOP=1 <<SQL
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname IN ('${SOURCE_DB}', '${TARGET_DB}')
  AND pid <> pg_backend_pid();

DROP DATABASE IF EXISTS ${TARGET_DB};
CREATE DATABASE ${TARGET_DB};
SQL

"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db pg_dump -U "${DB_USER}" -Fc -d "${SOURCE_DB}" > "${DUMP_FILE}"
"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db pg_restore -U "${DB_USER}" -d "${TARGET_DB}" --no-owner --no-privileges < "${DUMP_FILE}"
rm -f "${DUMP_FILE}"

echo "Copia completada en ${TARGET_DB}."
