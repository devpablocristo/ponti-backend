#!/usr/bin/env bash
set -euo pipefail

# Genera snapshot del schema
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SNAPSHOT_FILE="${ROOT_DIR}/scripts/db/schema.snapshot.sql"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/lib/backend_env.sh"
load_backend_env "${ROOT_DIR}"

echo "Generando snapshot de schema..."
"${ROOT_DIR}/scripts/compose_with_env.sh" exec -T ponti-db \
  pg_dump -U "${DB_USER}" -d "${DB_NAME}" --schema-only --no-owner --no-privileges \
  | sed -E '/^\\restrict /d; /^\\unrestrict /d' > "${SNAPSHOT_FILE}"

echo "Snapshot generado en ${SNAPSHOT_FILE}"
