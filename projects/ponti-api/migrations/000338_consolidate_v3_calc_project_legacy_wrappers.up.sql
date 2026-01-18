-- ========================================
-- MIGRACIÓN 000338: Consolidar v3_calc proyecto legacy -> v3_dashboard_ssot (UP)
-- ========================================
--
-- Propósito: Unificar funciones legacy en v3_calc para que deleguen en v3_dashboard_ssot.
-- Alcance: total_budget_cost_for_project, direct_costs_invested_for_project, stock_value_for_project.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Crear funciones SSOT en v3_dashboard_ssot (si no existían)
CREATE OR REPLACE FUNCTION v3_dashboard_ssot.total_budget_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  -- Placeholder histórico: 10x admin_cost
  SELECT COALESCE(p.admin_cost * 10, 0)::numeric
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.direct_costs_invested_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Labores invertidas (ejecutadas + no ejecutadas)
    (SELECT COALESCE(SUM(lb.price * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     JOIN public.labors lb ON lb.project_id = f.project_id AND lb.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Insumos invertidos (stock inicial)
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    -- Insumos recibidos por movimientos internos
    v3_dashboard_ssot.supply_cost_received_for_project(p_project_id)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_dashboard_ssot.stock_value_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Stock disponible = insumos comprados - insumos consumidos
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id 
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Wrappers en v3_calc
CREATE OR REPLACE FUNCTION v3_calc.total_budget_cost_for_project(p_project_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.total_budget_cost_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.direct_costs_invested_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.direct_costs_invested_for_project(p_project_id)
$$;

CREATE OR REPLACE FUNCTION v3_calc.stock_value_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_dashboard_ssot.stock_value_for_project(p_project_id)
$$;

COMMIT;
