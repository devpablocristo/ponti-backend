-- ========================================
-- MIGRACIÓN 000337: Consolidar v3_calc _mb -> v3_dashboard_ssot (DOWN)
-- ========================================
--
-- Propósito: Revertir wrappers y restaurar definiciones locales en v3_calc.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

CREATE OR REPLACE FUNCTION v3_calc.direct_costs_invested_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Insumos invertidos (stock inicial)
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    -- Labores planificadas (solo las que se ejecutaron)
    (SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::double precision
     FROM public.workorders w
     JOIN public.labors lb ON lb.id = w.labor_id
     JOIN public.lots l ON l.id = w.lot_id
     JOIN public.fields f ON f.id = l.field_id
     WHERE f.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.stock_value_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    v3_calc.direct_costs_invested_for_project_mb(p_project_id) - 
    v3_calc.direct_costs_total_for_project(p_project_id)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.agrochemicals_invested_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Stock inicial de agroquímicos
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND s.unit_id = 1 AND st.initial_units IS NOT NULL)
    +
    -- Movimientos de stock de agroquímicos
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Stock' 
       AND sm.deleted_at IS NULL AND s.unit_id = 1)
    +
    -- Movimientos de remito oficial de agroquímicos
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Remito oficial' 
       AND sm.deleted_at IS NULL AND s.unit_id = 1)
  , 0)::double precision
$$;

CREATE OR REPLACE FUNCTION v3_calc.seeds_invested_for_project_mb(p_project_id bigint)
RETURNS double precision
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    -- Stock inicial de semillas
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND s.unit_id = 2 AND st.initial_units IS NOT NULL)
    +
    -- Movimientos de stock de semillas
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Stock' 
       AND sm.deleted_at IS NULL AND s.unit_id = 2)
    +
    -- Movimientos de remito oficial de semillas
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE s.project_id = p_project_id AND sm.movement_type = 'Remito oficial' 
       AND sm.deleted_at IS NULL AND s.unit_id = 2)
  , 0)::double precision
$$;

COMMIT;
