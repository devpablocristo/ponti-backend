#!/usr/bin/env bash
set -euo pipefail

# Adopta baseline en un ambiente existente.
# Requiere: variables de conexión y DB ya existente.

if [[ $# -lt 2 ]]; then
  echo "Uso: $0 <db_host> <db_name> [db_user] [db_port] [sslmode]"
  exit 1
fi

DB_HOST="$1"
DB_NAME="$2"
DB_USER="${3:-postgres}"
DB_PORT="${4:-5432}"
DB_SSL_MODE="${5:-require}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASELINE_FILE="${ROOT_DIR}/migrations_v4/000000_baseline_schema.up.sql"

echo "Aplicando baseline (solo para ambientes existentes)..."
psql "host=${DB_HOST} port=${DB_PORT} user=${DB_USER} dbname=${DB_NAME} sslmode=${DB_SSL_MODE}" \
  -v ON_ERROR_STOP=1 -f "${BASELINE_FILE}"

echo "Marcando versión 000000 como aplicada..."
migrate -path "${ROOT_DIR}/migrations_v4" \
  -database "postgres://${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
  force 0

echo "Baseline adoptado."
