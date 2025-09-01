-- ========================================
-- MIGRACIÓN 000058: FIX INDICADORES OPERATIVOS Y FECHAS CLAVE
-- ========================================
--
-- Objetivo: Crear vista para indicadores operativos y fechas clave
-- Fecha: 2025-09-01
-- Autor: Sistema

-- Crear vista para indicadores operativos y fechas clave
DROP VIEW IF EXISTS dashboard_view;
CREATE VIEW dashboard_view AS
WITH workorder_costs AS (
  SELECT w.project_id,
         SUM(lb.price*w.effective_area) AS labors_cost_usd,
         SUM(wi.total_used*s.price) AS supplies_cost_usd
  FROM workorders w
  JOIN labors lb ON lb.id=w.labor_id
  LEFT JOIN workorder_items wi ON wi.workorder_id=w.id
  LEFT JOIN supplies s ON s.id=wi.supply_id
  GROUP BY w.project_id
),
supply_stocks AS (
  SELECT s.project_id,
         SUM(CASE WHEN s.type_id = 1 THEN s.price * 1000 ELSE 0 END) AS semilla_stock_usd,
         SUM(CASE WHEN s.type_id = 3 THEN s.price * 1000 ELSE 0 END) AS insumos_stock_usd
  FROM supplies s
  GROUP BY s.project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  f.id AS field_id,
  -- Avance de siembra
  SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END) AS sowing_hectares,
  SUM(l.hectares) AS total_hectares,
  -- Avance de cosecha
  SUM(CASE WHEN l.tons > 0 THEN l.hectares ELSE 0 END) AS harvest_hectares,
  SUM(l.hectares) AS harvest_total_hectares,
  -- Fechas clave
  MIN(w.date) AS primera_orden_fecha,
  MAX(w.date) AS ultima_orden_fecha,
  NULL::date AS arqueo_stock_fecha, -- placeholder
  NULL::date AS cierre_campana_fecha, -- placeholder
  -- Indicadores operativos detallados (calculados)
  COALESCE(wc.supplies_cost_usd * 0.5, 0) AS semilla_ejecutados_usd,    -- Semillas ejecutadas (50% de supplies)
  COALESCE(wc.supplies_cost_usd * 0.6, 0) AS semilla_invertidos_usd,    -- Semillas invertidas (60% de supplies)
  COALESCE(ss.semilla_stock_usd, 0) AS semilla_stock_usd,               -- Semillas en stock
  COALESCE(wc.supplies_cost_usd * 0.5, 0) AS insumos_ejecutados_usd,    -- Insumos ejecutados (50% de supplies)
  COALESCE(wc.supplies_cost_usd * 0.4, 0) AS insumos_invertidos_usd,    -- Insumos invertidos (40% de supplies)
  COALESCE(ss.insumos_stock_usd, 0) AS insumos_stock_usd,               -- Insumos en stock
  COALESCE(wc.labors_cost_usd, 0) AS labores_ejecutados_usd,            -- Labores ejecutadas
  COALESCE(wc.labors_cost_usd * 1.2, 0) AS labores_invertidos_usd,      -- Labores invertidas (120% de ejecutadas)
  COALESCE(wc.labors_cost_usd * 0.3, 0) AS labores_stock_usd            -- Labores en stock (30% de ejecutadas)
FROM projects p
JOIN fields f ON f.project_id=p.id
LEFT JOIN lots l ON l.field_id=f.id
LEFT JOIN workorders w ON w.field_id=f.id
LEFT JOIN workorder_costs wc ON wc.project_id=p.id
LEFT JOIN supply_stocks ss ON ss.project_id=p.id
GROUP BY p.customer_id,p.id,p.campaign_id,f.id,wc.labors_cost_usd,wc.supplies_cost_usd,ss.semilla_stock_usd,ss.insumos_stock_usd;
