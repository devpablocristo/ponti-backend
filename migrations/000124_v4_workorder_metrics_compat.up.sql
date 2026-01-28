-- ==========================================================
-- MIGRATION 000124 V4 WORKORDER METRICS COMPAT (UP)
-- ==========================================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Métricas de workorders con lógica compatible con v4
CREATE OR REPLACE VIEW v4_calc.workorder_metrics AS
WITH base AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.lot_id,
    w.effective_area,
    lb.price AS labor_price
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
labor_costs AS (
  SELECT
    project_id, field_id, lot_id,
    SUM((labor_price * effective_area))::numeric AS labor_cost_usd
  FROM base
  GROUP BY project_id, field_id, lot_id
),
supply_metrics AS (
  SELECT
    b.project_id, b.field_id, b.lot_id,
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS liters,
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * b.effective_area) ELSE 0 END)::numeric AS kilograms,
    SUM(v4_core.supply_cost(
      wi.final_dose::double precision,
      s.price::numeric,
      b.effective_area::numeric
    ))::numeric AS supplies_cost_usd
  FROM base b
  LEFT JOIN public.workorder_items wi
    ON wi.workorder_id = b.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s
    ON s.id = wi.supply_id AND s.deleted_at IS NULL
  GROUP BY b.project_id, b.field_id, b.lot_id
)
SELECT
  COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
  COALESCE(sur.field_id, lc.field_id, sm.field_id) AS field_id,
  COALESCE(sur.lot_id, lc.lot_id, sm.lot_id) AS lot_id,
  COALESCE(sur.surface_ha, 0)::numeric AS surface_ha,
  COALESCE(sm.liters, 0)::numeric AS liters,
  COALESCE(sm.kilograms, 0)::numeric AS kilograms,
  COALESCE(lc.labor_cost_usd, 0)::numeric AS labor_cost_usd,
  COALESCE(sm.supplies_cost_usd, 0)::numeric AS supplies_cost_usd,
  (COALESCE(lc.labor_cost_usd, 0)::numeric +
   COALESCE(sm.supplies_cost_usd, 0)::numeric) AS direct_cost_usd,
  v4_core.cost_per_ha(
    COALESCE(lc.labor_cost_usd, 0)::numeric + COALESCE(sm.supplies_cost_usd, 0)::numeric,
    COALESCE(sur.surface_ha, 0)::numeric
  ) AS avg_cost_per_ha_usd,
  v4_core.per_ha(COALESCE(sm.liters, 0)::numeric, COALESCE(sur.surface_ha, 0)::numeric) AS liters_per_ha,
  v4_core.per_ha(COALESCE(sm.kilograms, 0)::numeric, COALESCE(sur.surface_ha, 0)::numeric) AS kilograms_per_ha
FROM surface sur
FULL JOIN labor_costs lc USING (project_id, field_id, lot_id)
FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id);

COMMIT;
