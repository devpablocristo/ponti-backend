-- ========================================
-- MIGRACIÓN 000354: FIX ADMIN COST TOTAL FOR DASHBOARD (UP)
-- ========================================
--
-- Propósito: Alinear admin_cost_total_for_project con el comportamiento del dashboard remoto
-- (admin_cost × total_hectares)
--
-- Nota: Código en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  -- Retorna admin_cost × total_hectares para el Dashboard (costo total del proyecto)
  SELECT COALESCE(
    (SELECT p.admin_cost * v4_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

COMMIT;
