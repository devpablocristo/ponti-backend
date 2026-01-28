-- ========================================
-- MIGRACIÓN 000336: Consolidar v3_calc proyecto -> v3_dashboard_ssot (DOWN)
-- ========================================
--
-- Propósito: Revertir wrappers y restaurar definiciones locales en v3_calc.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Aggregations por proyecto
CREATE OR REPLACE FUNCTION v3_calc.total_hectares_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::double precision
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;

CREATE OR REPLACE FUNCTION v3_calc.total_invested_cost_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Arriendo invertido
    (SELECT COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Estructura invertida
    (SELECT COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.operating_result_total_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_calc.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costo administrativo total
    (SELECT COALESCE(p.admin_cost, 0)::double precision
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.total_costs_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados para todo el proyecto
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

-- Movimientos internos / costos de insumos por proyecto
CREATE OR REPLACE FUNCTION v3_calc.supply_cost_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders)
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.project_id = p_project_id)
    +
    -- Costos por movimientos internos de salida
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost_received_for_project(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos por movimientos internos de entrada (insumos recibidos de otros proyectos)
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno entrada'
       AND sm.is_entry = true
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;

-- Costos por cultivo
CREATE OR REPLACE FUNCTION v3_calc.total_costs_for_crop(p_project_id bigint, p_crop_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados para el cultivo
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT v3_calc.per_ha_dp(
    v3_calc.total_costs_for_crop(p_project_id, p_crop_id),
    (SELECT COALESCE(SUM(l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;

COMMIT;
