#!/usr/bin/env bash
set -euo pipefail

# Genera snapshot del schema
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"
SNAPSHOT_FILE="${ROOT_DIR}/scripts/db/schema.snapshot.sql"

set -a
source "${ENV_FILE}"
set +a

echo "Generando snapshot de schema..."
docker compose -f "${COMPOSE_FILE}" exec -T ponti-db \
  pg_dump -U "${DB_USER}" -d "${DB_NAME}" --schema-only --no-owner --no-privileges \
  | sed -E '/^\\restrict /d; /^\\unrestrict /d' > "${SNAPSHOT_FILE}"

echo "Snapshot generado en ${SNAPSHOT_FILE}"
