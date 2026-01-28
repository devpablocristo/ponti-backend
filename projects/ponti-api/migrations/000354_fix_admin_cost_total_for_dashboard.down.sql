-- ========================================
-- MIGRACIÓN 000354: FIX ADMIN COST TOTAL FOR DASHBOARD (DOWN)
-- ========================================
--
-- Revierte a admin_cost total sin multiplicar
--
-- Nota: Código en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(p.admin_cost, 0)::double precision
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

COMMIT;
