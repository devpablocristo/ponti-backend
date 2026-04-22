#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

echo "Asegurando base de datos ${DB_NAME}..."
"${ROOT_DIR}/scripts/compose_with_env.sh" up -d ponti-db >/dev/null

for i in {1..30}; do
  if "${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db pg_isready -U "${DB_USER}" -d postgres -p 5432 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

db_exists="$("${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db psql -U "${DB_USER}" -d postgres -At -c "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | tr -d '[:space:]' || true)"
if [[ "${db_exists}" == "1" ]]; then
  echo "Base de datos ${DB_NAME} ya existe."
  exit 0
fi

"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db psql -U "${DB_USER}" -d postgres -v ON_ERROR_STOP=1 <<SQL
CREATE DATABASE ${DB_NAME};
SQL

echo "Base de datos ${DB_NAME} creada."
