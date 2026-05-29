#!/usr/bin/env bash
set -euo pipefail

# Ejecuta migraciones v2 usando contenedor migrate
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"

set -a
source "${ENV_FILE}"
set +a

MIGRATIONS_PATH="${ROOT_DIR}/migrations_v4"
MIGRATIONS_RUN_PATH="$(mktemp -d)"

cp "${MIGRATIONS_PATH}"/*.sql "${MIGRATIONS_RUN_PATH}/"
rm -f "${MIGRATIONS_RUN_PATH}/000000_baseline_schema.up.sql" \
      "${MIGRATIONS_RUN_PATH}/000000_baseline_schema.down.sql"

echo "Aplicando migraciones v4..."
docker run --rm \
  --network "${ROOT_DIR##*/}_app-network" \
  -v "${MIGRATIONS_RUN_PATH}:/migrations" \
  migrate/migrate:v4.17.1 \
  -path /migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@ponti-db:5432/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
  up

rm -rf "${MIGRATIONS_RUN_PATH}"
