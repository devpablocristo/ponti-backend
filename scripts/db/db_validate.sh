#!/usr/bin/env bash
set -euo pipefail

# Ejecuta validaciones de esquema
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"
VALIDATE_SQL="${ROOT_DIR}/scripts/db/db_validate.sql"

set -a
source "${ENV_FILE}"
set +a

echo "Validando esquema..."
docker compose -f "${COMPOSE_FILE}" exec -T ponti-db \
  psql -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 < "${VALIDATE_SQL}"
