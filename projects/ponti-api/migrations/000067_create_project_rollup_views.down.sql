-- =====================================================
-- 000067: PROJECT - Vistas de Consolidación de Proyectos (ROLLBACK)
-- =====================================================
-- Entidad: project (Proyectos)
-- Funcionalidad: Eliminar vistas de consolidación de proyectos
-- =====================================================

-- Eliminar vistas de consolidación de proyectos
DROP VIEW IF EXISTS v_calc_project_costs CASCADE;
DROP VIEW IF EXISTS v_calc_project_economics CASCADE;
