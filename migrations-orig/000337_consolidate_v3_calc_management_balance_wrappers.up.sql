-- ========================================
-- MIGRACIÓN 000337: Consolidar v3_calc _mb -> v3_dashboard_ssot (UP)
-- ========================================
--
-- Propósito: Unificar funciones _mb en v3_calc para que deleguen en v3_dashboard_ssot.
-- Enfoque: v3_calc queda como wrapper de v3_dashboard_ssot (SSOT único para MB).
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

CREATE OR REPLACE FUNCTION v3_calc.seeds_invested_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.seeds_invested_for_project_mb(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.agrochemicals_invested_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.direct_costs_invested_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.direct_costs_invested_for_project_mb(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.stock_value_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.stock_value_for_project_mb(p_project_id)
$$;

COMMIT;
