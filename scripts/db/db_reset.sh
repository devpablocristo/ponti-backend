#!/usr/bin/env bash
set -euo pipefail

# Reset de DB local en contenedor
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

echo "Levantando DB..."
"${ROOT_DIR}/scripts/compose_with_env.sh" up -d ponti-db

echo "Esperando DB disponible..."
for i in {1..30}; do
  if "${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db pg_isready -U "${DB_USER}" -d "postgres" -p 5432 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "Recreando base de datos ${DB_NAME}..."
"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db psql -U "${DB_USER}" -d "postgres" -v ON_ERROR_STOP=1 <<SQL
-- Evitar "database is being accessed by other users" cuando quedó alguna sesión colgada
-- (ej. backend local, UI, pgadmin, etc.).
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '${DB_NAME}'
  AND pid <> pg_backend_pid();

DROP DATABASE IF EXISTS ${DB_NAME};
CREATE DATABASE ${DB_NAME};
SQL
