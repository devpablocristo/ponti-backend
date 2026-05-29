#!/usr/bin/env bash
set -euo pipefail

# Compara snapshot con expected
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
EXPECTED_FILE="${ROOT_DIR}/scripts/db/schema.expected.sql"
SNAPSHOT_FILE="${ROOT_DIR}/scripts/db/schema.snapshot.sql"

if [[ ! -f "${EXPECTED_FILE}" ]]; then
  echo "Falta schema.expected.sql. Generando uno inicial desde snapshot..."
  cp "${SNAPSHOT_FILE}" "${EXPECTED_FILE}"
  echo "schema.expected.sql creado. Revisalo y versionalo."
  exit 0
fi

if [[ ! -s "${EXPECTED_FILE}" ]]; then
  echo "schema.expected.sql vacío. Generando desde snapshot..."
  cp "${SNAPSHOT_FILE}" "${EXPECTED_FILE}"
  exit 0
fi

if [[ $(wc -c < "${EXPECTED_FILE}") -lt 200 ]]; then
  echo "schema.expected.sql es placeholder. Generando desde snapshot..."
  cp "${SNAPSHOT_FILE}" "${EXPECTED_FILE}"
  exit 0
fi

if [[ ! -f "${SNAPSHOT_FILE}" ]]; then
  echo "Falta schema.snapshot.sql. Ejecutá db_schema_snapshot.sh primero."
  exit 1
fi

echo "Comparando snapshot vs expected..."
diff -u "${EXPECTED_FILE}" "${SNAPSHOT_FILE}"
echo "Diff OK."
