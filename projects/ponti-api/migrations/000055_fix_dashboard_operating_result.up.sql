-- Migración 000055: Corregir SOLO el resultado operativo
-- Enfoque: Vista que calcula correctamente el resultado operativo
-- Partir de la migración 000050 (no tocar la 000051, 000052, 000053 ni 000054)

DROP VIEW IF EXISTS dashboard_operating_result_view;

CREATE OR REPLACE VIEW dashboard_operating_result_view AS
WITH labors_costs AS (
  SELECT
    w.project_id,
    SUM(lb.price * w.effective_area) AS executed_labors_usd
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.effective_area > 0
  GROUP BY w.project_id
),
supplies_costs AS (
  SELECT
    w.project_id,
    SUM(wi.total_used * s.price) AS executed_supplies_usd
  FROM workorders w
  JOIN workorder_items wi ON wi.workorder_id = w.id
  JOIN supplies s ON s.id = wi.supply_id
  WHERE wi.final_dose > 0
  GROUP BY w.project_id
),
harvest AS (
  SELECT f.project_id, SUM(l.tons) AS total_tons
  FROM fields f
  JOIN lots l ON l.field_id=f.id
  GROUP BY f.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  COALESCE(h.total_tons,0) AS total_tons,
  0::numeric AS income_usd, -- placeholder (lo calcula la app: tons*precio)
  COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0) AS operating_result_total_costs_usd,
  (0::numeric) - (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0)) AS operating_result_usd,
  CASE WHEN (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0)) > 0
       THEN ((0::numeric) - (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0))) / (COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0)) * 100
       ELSE 0 END AS operating_result_pct,
  COALESCE(lc.executed_labors_usd,0) + COALESCE(sc.executed_supplies_usd,0) AS executed_costs_usd
FROM projects p
LEFT JOIN labors_costs lc ON lc.project_id = p.id
LEFT JOIN supplies_costs sc ON sc.project_id = p.id
LEFT JOIN harvest h ON h.project_id = p.id;
