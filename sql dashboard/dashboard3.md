1. Avance de Siembra

-- Migración 000051: Corregir SOLO el avance de siembra
-- Enfoque: Vista simplificada que calcula correctamente el porcentaje de siembra

DROP VIEW IF EXISTS dashboard_view;

CREATE OR REPLACE VIEW dashboard_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    f.id AS field_id,
    SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0) AS sowing_hectares,
    SUM(l.hectares) AS sowing_total_hectares,
    CASE 
        WHEN SUM(l.hectares) > 0 THEN
            (SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0)::numeric / SUM(l.hectares) * 100)
        ELSE 0 
    END AS sowing_progress_percent
FROM projects p
JOIN fields f ON f.project_id = p.id
JOIN lots l ON l.field_id = f.id
GROUP BY p.customer_id, p.id, p.campaign_id, f.id;

2. Avance de Costos

-- Migración 000052: Corregir SOLO el avance de costos
-- Enfoque: Vista que calcula correctamente el porcentaje de avance de costos
-- Partir de la migración 000050 (no tocar la 000051)

DROP VIEW IF EXISTS dashboard_view;

CREATE OR REPLACE VIEW dashboard_view AS
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
    f.id AS field_id,
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
JOIN fields f ON f.project_id=p.id
LEFT JOIN costs c ON c.project_id=p.id;

