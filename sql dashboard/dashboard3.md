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

3. Avance de Cosecha

DROP VIEW IF EXISTS dashboard_view;

CREATE OR REPLACE VIEW dashboard_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    f.id AS field_id,
    COALESCE(SUM(CASE WHEN l.tons > 0 THEN l.hectares ELSE 0 END),0) AS harvest_hectares,
    COALESCE(SUM(l.hectares),0) AS harvest_total_hectares,
    CASE WHEN SUM(l.hectares)>0
         THEN (SUM(CASE WHEN l.tons > 0 THEN l.hectares ELSE 0 END)::numeric / SUM(l.hectares)::numeric * 100)
    ELSE 0 END AS harvest_progress_percent
FROM projects p
JOIN fields f ON f.project_id=p.id
JOIN lots l   ON l.field_id=f.id
GROUP BY p.customer_id,p.id,p.campaign_id,f.id;

4. Avance de Aportes

DROP VIEW IF EXISTS dashboard_view;

CREATE OR REPLACE VIEW dashboard_view AS
SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    f.id AS field_id,
    
    -- Campos de siembra (placeholder)
    0::numeric(14,2) AS sowing_hectares,
    0::numeric(14,2) AS sowing_total_hectares,
    0::numeric(6,2) AS sowing_progress_percent,
    
    -- Campos de cosecha (placeholder)
    0::numeric(14,2) AS harvest_hectares,
    0::numeric(14,2) AS harvest_total_hectares,
    0::numeric(6,2) AS harvest_progress_percent,
    
    -- Campos de costos (placeholder)
    0::numeric(14,2) AS executed_costs_usd,
    0::numeric(14,2) AS executed_labors_usd,
    0::numeric(14,2) AS executed_supplies_usd,
    0::numeric(14,2) AS budget_cost_usd,
    0::numeric(14,2) AS budget_total_usd,
    0::numeric(6,2) AS costs_progress_pct,
    
    -- Campos de resultado operativo (placeholder)
    0::numeric(14,2) AS income_usd,
    0::numeric(14,2) AS operating_result_usd,
    0::numeric(6,2) AS operating_result_pct,
    0::numeric(14,2) AS operating_result_total_costs_usd,
    
    -- Campos de aportes (REALES)
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    100.00::numeric(6,2) AS contributions_progress_pct,
    
    -- Campos de cultivos (placeholder)
    0::bigint AS crop_id,
    '' AS crop_name,
    0::numeric(14,2) AS crop_hectares,
    0::numeric(14,2) AS project_total_hectares,
    0::numeric(6,2) AS incidence_pct,
    0::numeric(14,2) AS crop_direct_costs_usd,
    0::numeric(14,2) AS cost_per_ha_usd,
    
    -- Campos de balance de gestión (placeholder)
    0::numeric(14,2) AS semilla_ejecutados_usd,
    0::numeric(14,2) AS semilla_invertidos_usd,
    0::numeric(14,2) AS semilla_stock_usd,
    0::numeric(14,2) AS insumos_ejecutados_usd,
    0::numeric(14,2) AS insumos_invertidos_usd,
    0::numeric(14,2) AS insumos_stock_usd,
    0::numeric(14,2) AS labores_ejecutados_usd,
    0::numeric(14,2) AS labores_invertidos_usd,
    0::numeric(14,2) AS labores_stock_usd,
    0::numeric(14,2) AS arriendo_ejecutados_usd,
    0::numeric(14,2) AS arriendo_invertidos_usd,
    0::numeric(14,2) AS arriendo_stock_usd,
    0::numeric(14,2) AS estructura_ejecutados_usd,
    0::numeric(14,2) AS estructura_invertidos_usd,
    0::numeric(14,2) AS estructura_stock_usd,
    0::numeric(14,2) AS costos_directos_ejecutados_usd,
    0::numeric(14,2) AS costos_directos_invertidos_usd,
    0::numeric(14,2) AS costos_directos_stock_usd,
    
    -- Campos de fechas (placeholder)
    NULL::timestamp AS primera_orden_fecha,
    0::bigint AS primera_orden_id,
    NULL::timestamp AS ultima_orden_fecha,
    0::bigint AS ultima_orden_id,
    NULL::timestamp AS arqueo_stock_fecha,
    NULL::timestamp AS cierre_campana_fecha,
    
    -- Campo de tipo de fila
    'metric'::text AS row_kind
    
FROM projects p
JOIN fields f ON f.project_id=p.id
JOIN project_investors pi ON pi.project_id=p.id
JOIN investors i ON i.id=pi.investor_id;

5. Resultado Operativo

DROP VIEW IF EXISTS dashboard_view;

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
