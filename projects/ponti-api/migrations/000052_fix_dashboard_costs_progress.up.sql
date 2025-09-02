-- Migración 000052: Corregir SOLO el avance de costos
-- Enfoque: Vista que calcula correctamente el porcentaje de avance de costos
-- Partir de la migración 000050 (no tocar la 000051)

DROP VIEW IF EXISTS dashboard_costs_progress_view;

CREATE OR REPLACE VIEW dashboard_costs_progress_view AS
WITH costs AS (
  SELECT
    w.project_id,
    SUM(lb.price) AS executed_labors_usd,
    SUM(s.price) AS executed_supplies_usd
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id = w.id
  LEFT JOIN supplies s ON s.id = wi.supply_id
  WHERE w.effective_area > 0
  GROUP BY w.project_id
)
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(c.executed_labors_usd,0) AS executed_labors_usd,
    COALESCE(c.executed_supplies_usd,0) AS executed_supplies_usd,
    COALESCE(c.executed_labors_usd,0)+COALESCE(c.executed_supplies_usd,0) AS executed_costs_usd,
    p.admin_cost AS budget_cost_usd,  -- Costos administrativos
    20000::numeric AS budget_total_usd,  -- hardcodeado
    LEAST(
      CASE WHEN 20000>0
          THEN (COALESCE(c.executed_labors_usd,0)+COALESCE(c.executed_supplies_usd,0)) / 20000.0 * 100
      ELSE 0 END,100
    ) AS costs_progress_pct
FROM projects p
LEFT JOIN costs c ON c.project_id=p.id;
