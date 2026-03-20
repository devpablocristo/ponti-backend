-- Down migration: UUID → bigint rollback
-- ADVERTENCIA: esta migración NO restaura los IDs originales.
-- Solo restaura los tipos de columna para compatibilidad de schema.
-- Los datos de mapping se pierden.

-- No implementada intencionalmente.
-- Si necesitás rollback, restaurá desde backup.
SELECT 1;
