#!/bin/bash
# Script para limpiar schemas de ramas/PRs
# Uso: ./cleanup_schema.sh pr_123
#   o: ./cleanup_schema.sh branch_feature_abc123

set -e

SCHEMA_NAME="${1:-}"

if [ -z "$SCHEMA_NAME" ]; then
    echo "Error: Schema name is required"
    echo "Usage: $0 <schema_name>"
    echo "Example: $0 pr_123"
    exit 1
fi

# Validaciones de seguridad
if [ "$SCHEMA_NAME" = "public" ]; then
    echo "Error: Cannot drop public schema"
    exit 1
fi

RESERVED_SCHEMAS=("pg_catalog" "pg_toast" "information_schema" "pg_temp" "pg_toast_temp")
for reserved in "${RESERVED_SCHEMAS[@]}"; do
    if [ "$SCHEMA_NAME" = "$reserved" ]; then
        echo "Error: Cannot drop reserved schema: $SCHEMA_NAME"
        exit 1
    fi
done

# Obtener variables de entorno de DB
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-ponti_api_db}"
DB_USER="${DB_USER:-admin}"
DB_PASSWORD="${DB_PASSWORD:-admin}"
DB_SSL_MODE="${DB_SSL_MODE:-disable}"

echo "Dropping schema: $SCHEMA_NAME"
echo "Database: $DB_NAME@$DB_HOST:$DB_PORT"

# Ejecutar DROP SCHEMA
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "DROP SCHEMA IF EXISTS \"$SCHEMA_NAME\" CASCADE;" \
    -v schema_name="$SCHEMA_NAME"

if [ $? -eq 0 ]; then
    echo "✅ Schema $SCHEMA_NAME dropped successfully"
else
    echo "❌ Failed to drop schema $SCHEMA_NAME"
    exit 1
fi
