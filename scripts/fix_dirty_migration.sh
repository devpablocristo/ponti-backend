#!/bin/bash
# Script para limpiar estado dirty de migraciones en un schema específico
# Uso: ./fix_dirty_migration.sh pr_5 79

set -e

SCHEMA_NAME="${1:-}"
VERSION="${2:-}"

if [ -z "$SCHEMA_NAME" ]; then
    echo "Error: Schema name is required"
    echo "Usage: $0 <schema_name> [version]"
    echo "Example: $0 pr_5 79"
    exit 1
fi

# Validaciones de seguridad
if [ "$SCHEMA_NAME" = "public" ]; then
    echo "Error: Cannot modify public schema migrations"
    exit 1
fi

# Obtener variables de entorno de DB
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-ponti_api_db}"
DB_USER="${DB_USER:-admin}"
DB_PASSWORD="${DB_PASSWORD:-admin}"
DB_SSL_MODE="${DB_SSL_MODE:-disable}"

echo "Fixing dirty migration state in schema: $SCHEMA_NAME"
echo "Database: $DB_NAME@$DB_HOST:$DB_PORT"
echo ""

# Ver estado actual
echo "Current migration state:"
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "SELECT version, dirty FROM ${SCHEMA_NAME}.schema_migrations ORDER BY version DESC LIMIT 5;" 2>/dev/null || {
    echo "Schema $SCHEMA_NAME does not exist or has no migrations table"
    exit 1
}

echo ""
echo "Fixing dirty state..."

if [ -n "$VERSION" ]; then
    # Forzar versión específica
    echo "Forcing version $VERSION to clean state..."
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -c "UPDATE ${SCHEMA_NAME}.schema_migrations SET dirty = false WHERE version = $VERSION;" 2>/dev/null || {
        echo "Failed to update version $VERSION"
        exit 1
    }
else
    # Limpiar todos los dirty
    echo "Cleaning all dirty migrations..."
    PGPASSWORD="$DB_PASSWORD" psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -c "UPDATE ${SCHEMA_NAME}.schema_migrations SET dirty = false WHERE dirty = true;" 2>/dev/null || {
        echo "Failed to clean dirty migrations"
        exit 1
    }
fi

echo ""
echo "✅ Dirty state fixed"
echo ""
echo "Updated migration state:"
PGPASSWORD="$DB_PASSWORD" psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -c "SELECT version, dirty FROM ${SCHEMA_NAME}.schema_migrations ORDER BY version DESC LIMIT 5;"
