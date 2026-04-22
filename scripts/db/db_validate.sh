#!/usr/bin/env bash
set -euo pipefail

# Ejecuta validaciones de esquema
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VALIDATE_SQL="${ROOT_DIR}/scripts/db/db_validate.sql"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

echo "Validando esquema..."
"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db \
  psql -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 < "${VALIDATE_SQL}"
