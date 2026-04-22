#!/usr/bin/env bash
set -euo pipefail

# Ejecuta migraciones v2 usando contenedor migrate
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

MIGRATIONS_PATH="${ROOT_DIR}/migrations_v4"
MIGRATIONS_RUN_PATH="$(mktemp -d)"

cp "${MIGRATIONS_PATH}"/*.sql "${MIGRATIONS_RUN_PATH}/"
rm -f "${MIGRATIONS_RUN_PATH}/000000_baseline_schema.up.sql" \
      "${MIGRATIONS_RUN_PATH}/000000_baseline_schema.down.sql"

echo "Aplicando migraciones v4..."
"${ROOT_DIR}/scripts/db/db_ensure_exists.sh" >/dev/null
docker run --rm \
  --network "${ROOT_DIR##*/}_app-network" \
  -v "${MIGRATIONS_RUN_PATH}:/migrations" \
  migrate/migrate:v4.17.1 \
  -path /migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@ponti-db:5432/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
  up

rm -rf "${MIGRATIONS_RUN_PATH}"
