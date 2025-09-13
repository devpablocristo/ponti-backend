-- ========================================
-- MIGRATION 000077: CREATE v3_workorder_metrics VIEW (UP)
-- ========================================
-- 
-- Purpose: Create aggregated metrics view by project/field/lot
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_workorder_metrics: métricas agregadas por proyecto/field/lot
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_workorder_metrics AS
WITH base AS (
  SELECT
    w.id                         AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.effective_area,
    lb.price                     AS labor_price
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
),
surface AS (
  SELECT project_id, field_id, lot_id, SUM(effective_area)::numeric AS surface_ha
  FROM base
  GROUP BY project_id, field_id, lot_id
),
-- Superficie al estilo v2 (suma con join a items/supplies que puede duplicar por item)
surface_v2 AS (
  SELECT
    w.project_id,
    w.field_id,
    w.lot_id,
    SUM(w.effective_area)::numeric AS surface_ha_v2
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi
         ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.project_id, w.field_id, w.lot_id
),
labor_costs AS (
  SELECT
    project_id, field_id, lot_id,
    SUM(public.calculate_labor_cost(labor_price, effective_area))::numeric AS labor_cost_usd
  FROM base
  GROUP BY project_id, field_id, lot_id
),
supply_metrics AS (
  SELECT
    b.project_id, b.field_id, b.lot_id,
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS liters,
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS kilograms,
    SUM(public.calculate_supply_cost(wi.final_dose::double precision,
                                     s.price::numeric,
                                     b.effective_area))::numeric          AS supplies_cost_usd
  FROM base b
  LEFT JOIN public.workorder_items wi
         ON wi.workorder_id = b.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
         ON s.id = wi.supply_id AND s.deleted_at IS NULL
  GROUP BY b.project_id, b.field_id, b.lot_id
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id, sv2.project_id) AS project_id,
  COALESCE(sur.field_id,  lc.field_id,  sm.field_id, sv2.field_id)       AS field_id,
  COALESCE(sur.lot_id,    lc.lot_id,    sm.lot_id, sv2.lot_id)           AS lot_id,
  COALESCE(sv2.surface_ha_v2, sur.surface_ha, 0)::numeric                AS surface_ha,
  COALESCE(sm.liters, 0)::numeric                        AS liters,
  COALESCE(sm.kilograms, 0)::numeric                     AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric                AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric             AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd, 0)::numeric +
   COALESCE(sm.supplies_cost_usd, 0)::numeric)           AS direct_cost_usd,
  public.calculate_cost_per_ha(
    COALESCE(lc.labor_cost_usd,0)::numeric + COALESCE(sm.supplies_cost_usd,0)::numeric,
    COALESCE(sur.surface_ha,0)::numeric
  )                                                       AS avg_cost_per_ha_usd,
  v3_calc.per_ha(COALESCE(sm.liters,0)::numeric, COALESCE(sur.surface_ha,0)::numeric)     AS liters_per_ha,
  v3_calc.per_ha(COALESCE(sm.kilograms,0)::numeric, COALESCE(sur.surface_ha,0)::numeric)  AS kilograms_per_ha
FROM surface sur
FULL JOIN labor_costs   lc USING (project_id, field_id, lot_id)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id)
FULL JOIN surface_v2   sv2 USING (project_id, field_id, lot_id);
