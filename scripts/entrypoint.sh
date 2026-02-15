#!/bin/sh
set -e

echo "=== Running migrations ==="
/app/migrate_binary || { echo "ERROR: migrations failed"; exit 1; }
echo "=== Migrations complete ==="

echo "=== Starting API server ==="
exec /app/prod_binary
