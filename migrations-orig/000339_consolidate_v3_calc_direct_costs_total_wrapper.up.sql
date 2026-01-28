-- ========================================
-- MIGRACIÓN 000339: Consolidar v3_calc.direct_costs_total_for_project (UP)
-- ========================================
--
-- Propósito: Unificar direct_costs_total_for_project en v3_calc como wrapper de v3_dashboard_ssot.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

CREATE OR REPLACE FUNCTION v3_calc.direct_costs_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.direct_costs_total_for_project(p_project_id)
$$;

COMMIT;
