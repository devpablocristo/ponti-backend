-- Script para limpiar schemas de ramas/PRs
-- Uso: psql -d ponti_api_db -f cleanup_schema.sql -v schema_name='pr_123'
-- O ejecutar desde Go/script con parámetros

-- Validar que el schema no sea 'public' ni un schema reservado
DO $$
BEGIN
    IF :'schema_name' = 'public' THEN
        RAISE EXCEPTION 'Cannot drop public schema';
    END IF;
    
    IF :'schema_name' IN ('pg_catalog', 'pg_toast', 'information_schema', 'pg_temp', 'pg_toast_temp') THEN
        RAISE EXCEPTION 'Cannot drop reserved schema: %', :'schema_name';
    END IF;
END $$;

-- Verificar que el schema existe antes de eliminarlo
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.schemata 
        WHERE schema_name = :'schema_name'
    ) THEN
        RAISE NOTICE 'Schema % does not exist, skipping', :'schema_name';
        RETURN;
    END IF;
END $$;

-- Eliminar schema con CASCADE (elimina todas las tablas, vistas, funciones, etc.)
DROP SCHEMA IF EXISTS :schema_name CASCADE;

-- Log de confirmación
DO $$
BEGIN
    RAISE NOTICE 'Schema % dropped successfully', :'schema_name';
END $$;
