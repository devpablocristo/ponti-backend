-- ========================================
-- MIGRACIÓN 000339: Consolidar v3_calc.direct_costs_total_for_project (DOWN)
-- ========================================
--
-- Propósito: Revertir wrapper y restaurar definición local en v3_calc.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

CREATE OR REPLACE FUNCTION v3_calc.direct_costs_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(wm.direct_cost_usd), 0)::double precision
     FROM public.v3_workorder_metrics wm
     WHERE wm.project_id = p_project_id)
  , 0)::double precision
$$;

COMMIT;
