-- ========================================
-- MIGRACIÓN 000336: Consolidar v3_calc proyecto -> v3_dashboard_ssot (UP)
-- ========================================
--
-- Propósito: Unificar funciones de proyecto en v3_calc para que deleguen en v3_dashboard_ssot.
-- Enfoque: v3_calc queda como wrapper de v3_dashboard_ssot (SSOT único para dashboard).
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Aggregations por proyecto
CREATE OR REPLACE FUNCTION v3_calc.total_hectares_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.total_hectares_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.total_invested_cost_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.total_invested_cost_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.operating_result_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.operating_result_total_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.total_costs_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.total_costs_for_project(p_project_id)
$$;

-- Movimientos internos / costos de insumos por proyecto
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.supply_cost_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost_received_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.supply_cost_received_for_project(p_project_id)
$$;

-- Costos por cultivo
CREATE OR REPLACE FUNCTION v3_calc.total_costs_for_crop(p_project_id bigint, p_crop_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.total_costs_for_crop(p_project_id, p_crop_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.cost_per_ha_for_crop(p_project_id, p_crop_id)
$$;

COMMIT;
