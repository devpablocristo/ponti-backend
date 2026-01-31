#!/usr/bin/env bash
set -euo pipefail

# Reset de DB local en contenedor
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"

set -a
source "${ENV_FILE}"
set +a

echo "Levantando DB..."
docker compose -f "${COMPOSE_FILE}" up -d ponti-db

echo "Esperando DB disponible..."
for i in {1..30}; do
  if docker compose -f "${COMPOSE_FILE}" exec -T ponti-db pg_isready -U "${DB_USER}" -d "postgres" -p "${DB_PORT}" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "Recreando base de datos ${DB_NAME}..."
docker compose -f "${COMPOSE_FILE}" exec -T ponti-db psql -U "${DB_USER}" -d "postgres" -v ON_ERROR_STOP=1 <<SQL
DROP DATABASE IF EXISTS ${DB_NAME};
CREATE DATABASE ${DB_NAME};
SQL
