-- ========================================
-- MIGRACIÓN 000056: FIX BALANCE DE GESTIÓN
-- ========================================
-- 
-- Objetivo: Crear vista para el balance de gestión
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Crear vista para el balance de gestión
DROP VIEW IF EXISTS dashboard_balance_management_view;
CREATE VIEW dashboard_balance_management_view AS
WITH base_costs AS (
  SELECT w.project_id,
         SUM(lb.price*w.effective_area) AS executed_labors_usd,
         SUM(wi.total_used*s.price) AS executed_supplies_usd
  FROM workorders w
  JOIN labors lb ON lb.id=w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id=w.id
  LEFT JOIN supplies s ON s.id=wi.supply_id
  GROUP BY w.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  COALESCE(bc.executed_labors_usd,0) AS executed_labors_usd,
  COALESCE(bc.executed_supplies_usd,0) AS executed_supplies_usd,
  COALESCE(bc.executed_labors_usd,0)+COALESCE(bc.executed_supplies_usd,0) AS balance_executed_costs_usd,
  p.admin_cost AS balance_budget_cost_usd,
  (COALESCE(bc.executed_labors_usd,0)+COALESCE(bc.executed_supplies_usd,0)+p.admin_cost) AS balance_operating_result_total_costs_usd,
  0::numeric AS balance_operating_result_usd, -- placeholder (lo calcula la app)
  0::numeric AS balance_operating_result_pct  -- placeholder (lo calcula la app)
FROM projects p
LEFT JOIN base_costs bc ON bc.project_id=p.id;
