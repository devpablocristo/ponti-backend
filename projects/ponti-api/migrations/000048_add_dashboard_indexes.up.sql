-- =========================================================
-- Agregar índices para mejorar el rendimiento del dashboard
-- =========================================================

-- Índices para mejorar JOINs en las vistas del dashboard
CREATE INDEX IF NOT EXISTS idx_fields_project_id ON fields(project_id);
CREATE INDEX IF NOT EXISTS idx_lots_field_id ON lots(field_id);
CREATE INDEX IF NOT EXISTS idx_lots_sowing_date ON lots(sowing_date);
CREATE INDEX IF NOT EXISTS idx_lots_tons ON lots(tons);
CREATE INDEX IF NOT EXISTS idx_labors_project_id ON labors(project_id);
CREATE INDEX IF NOT EXISTS idx_supplies_project_id ON supplies(project_id);
CREATE INDEX IF NOT EXISTS idx_project_investors_project_id ON project_investors(project_id);

-- Índices para filtros de deleted_at
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX IF NOT EXISTS idx_fields_deleted_at ON fields(deleted_at);
CREATE INDEX IF NOT EXISTS idx_lots_deleted_at ON lots(deleted_at);
CREATE INDEX IF NOT EXISTS idx_labors_deleted_at ON labors(deleted_at);
CREATE INDEX IF NOT EXISTS idx_supplies_deleted_at ON supplies(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_investors_deleted_at ON project_investors(deleted_at);
