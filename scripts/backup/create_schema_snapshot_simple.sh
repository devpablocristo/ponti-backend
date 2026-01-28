#!/bin/bash
# ========================================
# SCRIPT SIMPLE: Snapshot rápido del schema
# ========================================
# 
# Versión simplificada que solo exporta lo esencial

set -e

SNAPSHOT_NAME="${1:-snapshot_$(date +%Y%m%d_%H%M%S)}"
OUTPUT_FILE="snapshots/${SNAPSHOT_NAME}.sql"

mkdir -p snapshots

echo "📸 Creando snapshot rápido: ${SNAPSHOT_NAME}"

# Exportar solo schemas v4 y vistas v4_report
docker compose -f docker-compose.yml exec -T ponti-db \
    pg_dump -U admin -d ponti_api_db \
    --schema=v4_core \
    --schema=v4_ssot \
    --schema=v4_calc \
    --schema=v4_report \
    --schema-only \
    --no-owner \
    --no-privileges \
    > "${OUTPUT_FILE}"

# Agregar vistas v4_report al final
echo "" >> "${OUTPUT_FILE}"
echo "-- ========================================" >> "${OUTPUT_FILE}"
echo "-- VISTAS v4_report" >> "${OUTPUT_FILE}"
echo "-- ========================================" >> "${OUTPUT_FILE}"

docker compose -f docker-compose.yml exec -T ponti-db \
    psql -U admin -d ponti_api_db -t -c "
    SELECT 'CREATE OR REPLACE VIEW ' || schemaname || '.' || viewname || ' AS ' || 
           pg_get_viewdef(schemaname || '.' || viewname, true) || ';'
    FROM pg_views
    WHERE schemaname = 'v4_report'
    ORDER BY viewname;
    " >> "${OUTPUT_FILE}"

echo "✅ Snapshot guardado en: ${OUTPUT_FILE}"
echo "📊 Tamaño: $(du -h ${OUTPUT_FILE} | cut -f1)"





