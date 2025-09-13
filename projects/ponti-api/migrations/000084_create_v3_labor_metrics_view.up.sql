-- ========================================
-- MIGRATION 000082: CREATE v3_labor_metrics VIEW (UP)
-- ========================================
-- 
-- Purpose: Labor metrics by project/field
-- Date: 2025-09-12
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

-- -------------------------------------------------------------------
-- v3_labor_metrics: métricas agregadas de labores
-- -------------------------------------------------------------------
CREATE OR REPLACE VIEW public.v3_labor_metrics AS
WITH wo AS (
  SELECT
    w.id AS workorder_id,
    w.project_id,
    w.field_id,
    w.date,
    w.effective_area::numeric AS effective_area,
    lb.price::numeric AS labor_price_per_ha
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
),
agg AS (
  SELECT
    project_id,
    field_id,
    COUNT(DISTINCT workorder_id) AS total_workorders,
    SUM(effective_area) AS surface_ha,
    SUM(labor_price_per_ha * effective_area) AS total_labor_cost,
    MIN(date) AS first_workorder_date,
    MAX(date) AS last_workorder_date
  FROM wo
  GROUP BY project_id, field_id
)
SELECT
  a.project_id,
  a.field_id,
  a.surface_ha,
  a.total_labor_cost,
  v3_calc.cost_per_ha(a.total_labor_cost, a.surface_ha) AS avg_labor_cost_per_ha,
  a.total_workorders,
  a.first_workorder_date,
  a.last_workorder_date
FROM agg a;


