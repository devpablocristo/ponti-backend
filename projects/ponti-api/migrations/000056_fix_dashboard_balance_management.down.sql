-- ========================================
-- ROLLBACK MIGRACIÓN 000056: FIX BALANCE DE GESTIÓN
-- ========================================
-- 
-- Objetivo: Restaurar vista dashboard_view al estado anterior
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Restaurar vista dashboard_view a su estado anterior (migración 000055)
CREATE OR REPLACE VIEW dashboard_view AS
WITH costs AS (
  SELECT w.project_id,
         SUM(lb.price*w.effective_area) + SUM(wi.total_used*s.price) AS executed_costs_usd
  FROM workorders w
  JOIN labors lb ON lb.id=w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id=w.id
  LEFT JOIN supplies s ON s.id=wi.supply_id
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
  f.id AS field_id,
  COALESCE(h.total_tons,0) AS total_tons,
  0::numeric AS income_usd, -- placeholder (lo calcula la app: tons*precio)
  COALESCE(c.executed_costs_usd,0) AS total_invested_usd,
  0::numeric AS operating_result_usd, -- placeholder
  0::numeric AS operating_result_pct, -- placeholder
  COALESCE(c.executed_costs_usd,0) AS operating_result_total_costs_usd -- Costos totales ejecutados
FROM projects p
JOIN fields f ON f.project_id=p.id
LEFT JOIN costs c ON c.project_id=p.id
LEFT JOIN harvest h ON h.project_id=p.id;
