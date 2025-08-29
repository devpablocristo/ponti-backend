-- =========================================================
-- Rollback: Eliminar índices del dashboard
-- =========================================================

-- Eliminar índices de JOINs
DROP INDEX IF EXISTS idx_fields_project_id;
DROP INDEX IF EXISTS idx_lots_field_id;
DROP INDEX IF EXISTS idx_lots_sowing_date;
DROP INDEX IF EXISTS idx_lots_tons;
DROP INDEX IF EXISTS idx_labors_project_id;
DROP INDEX IF EXISTS idx_supplies_project_id;
DROP INDEX IF EXISTS idx_project_investors_project_id;

-- Eliminar índices de deleted_at
DROP INDEX IF EXISTS idx_projects_deleted_at;
DROP INDEX IF EXISTS idx_fields_deleted_at;
DROP INDEX IF EXISTS idx_lots_deleted_at;
DROP INDEX IF EXISTS idx_labors_deleted_at;
DROP INDEX IF EXISTS idx_supplies_deleted_at;
DROP INDEX IF EXISTS idx_project_investors_deleted_at;
