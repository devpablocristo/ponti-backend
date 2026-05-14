#!/bin/sh
set -e

echo "=== Running migrations ==="
if [ "${RUN_MIGRATIONS_ON_STARTUP:-true}" = "true" ]; then
  /app/migrate_binary || { echo "ERROR: migrations failed"; exit 1; }
  echo "=== Migrations complete ==="
else
  echo "=== Migrations skipped (RUN_MIGRATIONS_ON_STARTUP=false) ==="
fi

echo "=== Starting API server ==="
exec /app/prod_binary
